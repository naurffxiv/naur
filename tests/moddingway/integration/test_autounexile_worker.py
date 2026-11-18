"""
The autounexile worker, run the way its loop runs it, with the real database.
"""

import logging
from datetime import timedelta

import discord

from moddingway.constants import ExileStatus
from moddingway.services import exiles
from moddingway.workers.autounexile import autounexile_users

from .conftest import EXILED_ROLE_ID, VERIFIED_ROLE_ID

COMPLETED = "Auto Unexile task completed."


def fields(embed: discord.Embed) -> list[tuple[str | None, str | None]]:
    return [(field.name, field.value) for field in embed.fields]


def statuses(query) -> list[int]:
    return [row[0] for row in query("SELECT exileStatus FROM exiles ORDER BY 1")]


async def exile_ended(member) -> exiles.ExileResult:
    return await exiles.exile(member, timedelta(seconds=-1), "already over")


async def test_worker__releases_due_exile(bot, make_member, log_channel, query):
    member = make_member()
    exiled = await exile_ended(member)

    assert await autounexile_users(bot) == COMPLETED

    assert statuses(query) == [ExileStatus.UNEXILED]
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    [embed] = log_channel.embeds
    assert embed.description == f"<@{member.id}>'s exile has timed out"
    assert embed.footer.text == f"Exile ID: {exiled.exile_id}"
    assert ("Result", f"<@{member.id}> was successfully unexiled") in fields(embed)


async def test_worker__nothing_due(bot, make_member, log_channel, query):
    await exiles.exile(make_member(), timedelta(days=1), "reason")

    assert await autounexile_users(bot) == COMPLETED

    assert statuses(query) == [ExileStatus.TIMED_EXILED]
    assert log_channel.embeds == []


async def test_worker__missing_guild_is_skipped(
    bot, make_member, log_channel, query, caplog
):
    await exile_ended(make_member())
    bot.guild = None

    with caplog.at_level(logging.ERROR):
        assert await autounexile_users(bot) is None

    assert "not found, skipping auto unexile" in caplog.text
    assert statuses(query) == [ExileStatus.TIMED_EXILED]


async def test_worker__member_gone_ends_exile(
    bot, guild, make_member, log_channel, query
):
    member = make_member()
    await exile_ended(member)
    del guild.members[member.id]

    assert await autounexile_users(bot) == COMPLETED

    assert statuses(query) == [ExileStatus.UNEXILED]
    [embed] = log_channel.embeds
    [(name, value)] = fields(embed)
    assert name == "Result"
    assert value is not None and "was not found in the server" in value


async def test_worker__failure_is_logged_and_retried(
    bot, make_member, log_channel, query, caplog
):
    member = make_member()
    await exile_ended(member)
    member.failing_role_ids.add(VERIFIED_ROLE_ID)

    with caplog.at_level(logging.INFO):
        assert await autounexile_users(bot) == COMPLETED

    assert "it stays active and will be retried" in caplog.text
    assert statuses(query) == [ExileStatus.TIMED_EXILED]
    assert member.role_ids() == {EXILED_ROLE_ID}
    [embed] = log_channel.embeds
    [(name, value)] = fields(embed)
    assert name == "Error"
    assert value is not None
    assert value.startswith("Unexile failed and no changes were saved")

    member.failing_role_ids.clear()
    assert await autounexile_users(bot) == COMPLETED
    assert statuses(query) == [ExileStatus.UNEXILED]


async def test_worker__logging_failure_does_not_stop_the_run(
    bot, make_member, log_channel, query, caplog
):
    members = [make_member(member_id=1), make_member(member_id=2)]
    for member in members:
        await exile_ended(member)
    log_channel.fails = True

    assert await autounexile_users(bot) == COMPLETED

    assert statuses(query) == [ExileStatus.UNEXILED] * len(members)
    assert caplog.text.count("Failed to log auto unexile") == len(members)
