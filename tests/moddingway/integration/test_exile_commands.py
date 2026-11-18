"""
The exile slash commands, driven through the callbacks they register, with a
fake interaction and the real database.
"""

from datetime import timedelta
from typing import Any

import discord
import pytest

from moddingway.commands import exile_commands
from moddingway.constants import ExileStatus, Role
from moddingway.services import exiles
from moddingway.services.exile_service import (
    MISSING_USER_RECORD_MESSAGE,
    NOT_EXILED_MESSAGE,
    NOT_VERIFIED_MESSAGE,
)

from .conftest import EXILED_ROLE_ID, MOD_ROLE_ID, VERIFIED_ROLE_ID, FakeInteraction

MODERATOR_ID = 999


def exile_rows(query) -> list[tuple]:
    return query(
        """
        SELECT e.exileID, u.discordUserID, e.exileStatus, e.reason
        FROM exiles e JOIN users u ON e.userID = u.userID
        ORDER BY e.exileID
        """
    )


def fields(embed: discord.Embed) -> list[tuple[str | None, str | None]]:
    return [(field.name, field.value) for field in embed.fields]


@pytest.fixture
def commands(bot) -> dict[str, Any]:
    exile_commands.create_exile_commands(bot)
    return bot.tree.commands


@pytest.fixture
def interaction(guild, make_member) -> Any:
    def __interaction(command_name: str | None, user=None) -> Any:
        if user is None:
            user = make_member(member_id=MODERATOR_ID, roles=[Role.MOD])
        return FakeInteraction(user, guild, command_name)

    return __interaction


# /exile


async def test_exile__member_is_exiled(
    commands, interaction, make_member, log_channel, query
):
    member = make_member()
    command = interaction("exile")

    await commands["exile"](command, member, "1day", "spamming")

    assert command.reply == f"Successfully exiled <@{member.id}>"
    [(exile_id, discord_id, status, reason)] = exile_rows(query)
    assert (discord_id, status, reason) == (
        str(member.id),
        ExileStatus.TIMED_EXILED,
        "spamming",
    )
    assert member.role_ids() == {EXILED_ROLE_ID}

    [embed] = log_channel.embeds
    assert ("Action", "/exile") in fields(embed)
    assert ("Duration", "1 day") in fields(embed)
    assert ("Result", f"<@{member.id}> was successfully exiled") in fields(embed)
    assert embed.footer.text == f"Exile ID: {exile_id}"


async def test_exile__exiled_member_is_extended(
    commands, interaction, make_member, log_channel, query
):
    member = make_member()
    await commands["exile"](interaction("exile"), member, "1day", "first")
    command = interaction("exile")

    await commands["exile"](command, member, "2day", "second")

    assert command.reply == "User exile extended"
    assert len(exile_rows(query)) == 1
    embed = log_channel.embeds[-1]
    [(_, value)] = [f for f in fields(embed) if f[0] == "Result"]
    assert value is not None and value.startswith("Exile extended until <t:")
    # Extensions have never shown the exile ID
    assert embed.footer.text is None


async def test_exile__unverified_member_is_refused(
    commands, interaction, make_member, log_channel, query
):
    member = make_member(roles=[Role.NON_VERIFIED])
    command = interaction("exile")

    await commands["exile"](command, member, "1day", "reason")

    assert command.reply == NOT_VERIFIED_MESSAGE
    assert exile_rows(query) == []
    assert ("Error", NOT_VERIFIED_MESSAGE) in fields(log_channel.embeds[0])


async def test_exile__role_swap_failure_saves_nothing(
    commands, interaction, make_member, log_channel, query
):
    member = make_member()
    member.failing_role_ids.add(EXILED_ROLE_ID)
    command = interaction("exile")

    await commands["exile"](command, member, "1day", "reason")

    assert command.reply is not None
    assert command.reply.startswith("Exile failed and no changes were saved")
    assert exile_rows(query) == []
    assert query("SELECT * FROM users") == []
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    assert ("Error", command.reply) in fields(log_channel.embeds[0])


async def test_exile__invalid_duration_is_refused(
    commands, interaction, make_member, log_channel, query
):
    command = interaction("exile")

    await commands["exile"](command, make_member(), "1week", "reason")

    [(message, kwargs)] = command.response.sent
    assert message.startswith("Invalid exile duration given")
    assert kwargs == {"ephemeral": True}
    assert exile_rows(query) == []
    assert log_channel.embeds == []


async def test_exile__empty_duration_is_refused(
    commands, interaction, make_member, log_channel, query
):
    command = interaction("exile")

    await commands["exile"](command, make_member(), "", "reason")

    assert command.reply == "Invalid exile duration."
    assert exile_rows(query) == []


async def test_exile__mod_cannot_be_exiled(
    commands, interaction, make_member, log_channel, query
):
    target = make_member(member_id=1, roles=[Role.VERIFIED, Role.MOD])
    command = interaction("exile")

    await commands["exile"](command, target, "1day", "reason")

    [(message, _)] = command.response.sent
    assert message == f"Unable to exile <@{target.id}>: You cannot exile a mod."
    assert exile_rows(query) == []
    assert log_channel.embeds == []


# /unexile


async def test_unexile__member_is_unexiled(
    commands, interaction, make_member, log_channel, query
):
    member = make_member()
    exiled = await exiles.exile(member, timedelta(days=1), "reason")
    command = interaction("unexile")

    await commands["unexile"](command, member)

    assert command.reply == f"Successfully unexiled <@{member.id}>"
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNEXILED
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    [embed] = log_channel.embeds
    assert ("Result", f"<@{member.id}> was successfully unexiled") in fields(embed)
    assert embed.footer.text == f"Exile ID: {exiled.exile_id}"


async def test_unexile__member_who_is_not_exiled_is_refused(
    commands, interaction, make_member, log_channel
):
    member = make_member()
    command = interaction("unexile")

    await commands["unexile"](command, member)

    assert command.reply == NOT_EXILED_MESSAGE
    assert ("Error", NOT_EXILED_MESSAGE) in fields(log_channel.embeds[0])
    assert member.role_ids() == {VERIFIED_ROLE_ID}


async def test_unexile__member_without_user_record(
    commands, interaction, make_member, log_channel, query
):
    member = make_member(roles=[Role.EXILED])
    command = interaction("unexile")

    await commands["unexile"](command, member)

    assert command.reply == MISSING_USER_RECORD_MESSAGE
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    assert ("Error", MISSING_USER_RECORD_MESSAGE) in fields(log_channel.embeds[0])


async def test_unexile__role_swap_failure_saves_nothing(
    commands, interaction, make_member, log_channel, query
):
    member = make_member()
    await exiles.exile(member, timedelta(days=1), "reason")
    member.failing_role_ids.add(VERIFIED_ROLE_ID)
    command = interaction("unexile")

    await commands["unexile"](command, member)

    assert command.reply is not None
    assert command.reply.startswith("Unexile failed and no changes were saved")
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.TIMED_EXILED
    assert member.role_ids() == {EXILED_ROLE_ID}


# /roulette


@pytest.fixture
def roulette(commands, interaction, monkeypatch):
    """
    Run /roulette and press a button. ``lose`` picks the losing chamber and the
    longest exile, otherwise the first (safe) option.
    """

    async def __roulette(player, *, lose: bool, button: str = "Confirm") -> Any:
        monkeypatch.setattr(
            exile_commands,
            "choice",
            lambda options: options[-1] if lose else options[0],
        )
        command = interaction("roulette", user=player)
        await commands["roulette"](command)
        [(_, kwargs)] = command.response.sent
        view = kwargs["view"]
        [pressed] = [item for item in view.children if item.label == button]

        button_press = interaction(None, user=player)
        await pressed.callback(button_press)
        assert command.edited_views == [view]
        return button_press

    return __roulette


async def test_roulette__survivor_is_not_exiled(
    roulette, make_member, log_channel, query
):
    player = make_member()

    press = await roulette(player, lose=False)

    [(message, _)] = press.response.sent
    assert message == f"<@{player.id}> has tested their luck and lives another day..."
    assert exile_rows(query) == []
    assert log_channel.embeds == []


async def test_roulette__loser_is_exiled(roulette, make_member, log_channel, query):
    player = make_member()

    press = await roulette(player, lose=True)

    assert press.reply == (
        f"<@{player.id}> has tested their luck and has utterly failed! "
        f"<@{player.id}> has been sent into exile for 24 hour(s)."
    )
    [(_, _, status, reason)] = exile_rows(query)
    assert (status, reason) == (ExileStatus.TIMED_EXILED, "lost roulette")
    assert player.role_ids() == {EXILED_ROLE_ID}
    [embed] = log_channel.embeds
    assert embed.description is not None
    assert embed.description.startswith(
        f"User <@{player.id}> lost a roulette in #general and was exiled."
    )
    assert ("Duration", "24 hours") in fields(embed)


async def test_roulette__losing_mod_is_only_told(
    roulette, make_member, log_channel, query
):
    player = make_member(roles=[Role.VERIFIED, Role.MOD])

    press = await roulette(player, lose=True)

    [(message, _)] = press.response.sent
    assert "has been sent into exile for 24 hour(s)" in message
    assert exile_rows(query) == []
    assert player.role_ids() == {VERIFIED_ROLE_ID, MOD_ROLE_ID}


async def test_roulette__failed_exile_is_reported(
    roulette, make_member, log_channel, query
):
    player = make_member()
    player.failing_role_ids.add(EXILED_ROLE_ID)

    press = await roulette(player, lose=True)

    assert press.reply == "An error occurred while processing the command."
    assert exile_rows(query) == []
    [embed] = log_channel.embeds
    assert any(name == "Error" for name, _ in fields(embed))


async def test_roulette__outside_a_server(roulette, guild, log_channel, query):
    player = discord.Object(id=1)

    press = await roulette(player, lose=True)

    [(message, _)] = press.response.sent
    assert message == "This command must be run in a server."
    assert exile_rows(query) == []


async def test_roulette__cancelled(roulette, make_member, query):
    press = await roulette(make_member(), lose=True, button="Cancel")

    assert press.reply == "Roulette cancelled"
    assert exile_rows(query) == []


# Exile history and the manual worker run


async def test_view_exiles_and_delete_exile(
    commands, interaction, make_member, log_channel, query
):
    member = make_member()
    exiled = await exiles.exile(member, timedelta(days=1), "spamming")

    active = interaction("view_active_exiles")
    await commands["view_active_exiles"](active)
    assert active.reply is not None
    assert f"ID: {exiled.exile_id} | USER: <@{member.id}>" in active.reply

    logged = interaction("view_logged_exiles")
    await commands["view_logged_exiles"](logged, member)
    assert logged.reply is not None and "REASON: spamming" in logged.reply

    await exiles.unexile(member)
    delete = interaction("delete_exile")
    await commands["delete_exile"](delete, exiled.exile_id)

    assert delete.reply == f"Exile {exiled.exile_id} successfully deleted"
    assert exile_rows(query) == []
    assert ("Exile Id", str(exiled.exile_id)) in fields(log_channel.embeds[-1])


async def test_run_auto_unexile(commands, interaction, make_member, log_channel, query):
    member = make_member()
    await exiles.exile(member, timedelta(seconds=-1), "already over")
    command = interaction("run_auto_unexile")

    await commands["run_auto_unexile"](command)

    assert command.reply == "Auto Unexile task completed."
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNEXILED
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    # One embed from the worker, one for the command
    assert ("Result", "Auto Unexile task completed.") in fields(log_channel.embeds[-1])
