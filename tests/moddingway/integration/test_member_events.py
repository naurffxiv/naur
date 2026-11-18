"""
Member events, driven through the handlers the bot registers, with the real
database.
"""

from datetime import UTC, datetime, timedelta
from typing import Any

import discord
import pytest

from moddingway.constants import ExileStatus
from moddingway.events.member_events import register_events
from moddingway.services import exiles

from .conftest import EXILED_ROLE_ID, NON_VERIFIED_ROLE_ID, STICKY_ROLE_IDS

STICKY_ROLE = STICKY_ROLE_IDS[0]


def fields(embed: discord.Embed) -> list[tuple[str | None, str | None]]:
    return [(field.name, field.value) for field in embed.fields]


def sticky_rows(query) -> list[str]:
    return [row[0] for row in query("SELECT roleID FROM roles ORDER BY entryID")]


@pytest.fixture
def events(bot) -> dict[str, Any]:
    register_events(bot)
    return bot.events


@pytest.fixture
def rejoin(make_member) -> Any:
    """The same member joining again, with no roles."""

    def __rejoin(member) -> Any:
        return make_member(member_id=member.id, roles=[])

    return __rejoin


# on_member_join


async def test_join__new_member_gets_non_verified(events, make_member, log_channel):
    member = make_member(roles=[])

    await events["on_member_join"](member)

    assert member.role_ids() == {NON_VERIFIED_ROLE_ID}
    [embed] = log_channel.embeds
    assert embed.description == f"<@{member.id}> joined the server"
    assert embed.footer.text == "Member Joined"
    assert fields(embed)[0] == (
        "Result",
        f"Successfully assigned <@&{NON_VERIFIED_ROLE_ID}> role",
    )
    assert fields(embed)[1][0] == "Account Age"


async def test_join__rejoin_during_exile_gets_exiled(
    events, make_member, rejoin, log_channel, query
):
    exiled = await exiles.exile(make_member(), timedelta(days=1), "reason")
    member = rejoin(exiled.member)

    await events["on_member_join"](member)

    assert member.role_ids() == {EXILED_ROLE_ID}
    [embed] = log_channel.embeds
    assert embed.footer.text == f"Exile ID: {exiled.exile_id}"
    [(_, result), (expiration, _)] = fields(embed)
    assert result is not None and "rejoined during an active exile" in result
    assert expiration == "Expiration"


async def test_join__rejoin_after_exile_gets_non_verified_and_clears_sticky_roles(
    events, guild, make_member, rejoin, log_channel, query
):
    member = make_member(role_ids=[STICKY_ROLE])
    await exiles.exile(member, timedelta(days=1), "reason")
    del guild.members[member.id]
    # The worker ends the exile while the member is away; sticky roles stay saved
    await exiles.release_due(guild, datetime.now(UTC) + timedelta(days=2))
    assert sticky_rows(query) == [str(STICKY_ROLE)]
    member = rejoin(member)

    await events["on_member_join"](member)

    assert member.role_ids() == {NON_VERIFIED_ROLE_ID}
    assert sticky_rows(query) == []
    [(_, status)] = query("SELECT exileID, exileStatus FROM exiles")
    assert status == ExileStatus.UNEXILED
    assert fields(log_channel.embeds[0])[0][0] == "Result"


async def test_join__failed_exile_lookup_falls_back_to_non_verified(
    events, make_member, log_channel, monkeypatch
):
    async def lookup_fails(member):
        raise RuntimeError("database unavailable")

    monkeypatch.setattr(exiles, "reapply_on_rejoin", lookup_fails)
    member = make_member(roles=[])

    await events["on_member_join"](member)

    assert member.role_ids() == {NON_VERIFIED_ROLE_ID}
    assert fields(log_channel.embeds[0])[0][0] == "Result"


async def test_join__failed_reapply_does_not_give_non_verified(
    events, make_member, rejoin, log_channel
):
    exiled = await exiles.exile(make_member(), timedelta(days=1), "reason")
    member = rejoin(exiled.member)
    member.failing_role_ids.add(EXILED_ROLE_ID)

    await events["on_member_join"](member)

    assert member.roles == []
    [embed] = log_channel.embeds
    [(name, value)] = fields(embed)
    assert name == "Error"
    assert value is not None and "Exiled role could not be assigned" in value


async def test_join__failed_non_verified_is_logged(events, make_member, log_channel):
    member = make_member(roles=[])
    member.failing_role_ids.add(NON_VERIFIED_ROLE_ID)

    await events["on_member_join"](member)

    assert member.roles == []
    [(name, value)] = fields(log_channel.embeds[0])
    assert name == "Error"
    assert value is not None and value.startswith("Failed to assign role")


async def test_join__without_logging_channel_still_assigns_role(
    events, make_member, rejoin
):
    exiled = await exiles.exile(make_member(), timedelta(days=1), "reason")
    returning = rejoin(exiled.member)
    newcomer = make_member(member_id=2, roles=[])

    await events["on_member_join"](returning)
    await events["on_member_join"](newcomer)

    assert returning.role_ids() == {EXILED_ROLE_ID}
    assert newcomer.role_ids() == {NON_VERIFIED_ROLE_ID}


async def test_join__logging_failure_is_contained(events, make_member, log_channel):
    log_channel.fails = True
    member = make_member(roles=[])

    await events["on_member_join"](member)

    assert member.role_ids() == {NON_VERIFIED_ROLE_ID}


# on_member_ban / on_member_unban


async def test_ban_and_unban__are_saved_and_logged(
    events, guild, make_member, log_channel, query
):
    member = make_member()

    await events["on_member_ban"](guild, member)
    assert query("SELECT discordUserID, isBanned FROM users") == [
        (str(member.id), True)
    ]

    await events["on_member_unban"](guild, member)
    assert query("SELECT discordUserID, isBanned FROM users") == [
        (str(member.id), False)
    ]

    ban, unban = log_channel.embeds
    assert ban.description == f"<@{member.id}> was banned"
    assert ("Action", "/ban") in fields(ban)
    assert unban.description == f"<@{member.id}> was unbanned"
    assert ("Action", "/unban") in fields(unban)


async def test_ban_and_unban__without_logging_channel(
    events, guild, make_member, query
):
    member = make_member()

    await events["on_member_ban"](guild, member)
    assert query("SELECT isBanned FROM users") == [(True,)]
    await events["on_member_unban"](guild, member)
    assert query("SELECT isBanned FROM users") == [(False,)]


@pytest.mark.parametrize("event", ["on_member_ban", "on_member_unban"])
async def test_ban_and_unban__logging_failure_is_contained(
    events, guild, make_member, log_channel, query, event
):
    log_channel.fails = True

    await events[event](guild, make_member())

    assert log_channel.embeds == []
    assert len(query("SELECT * FROM users")) == 1
