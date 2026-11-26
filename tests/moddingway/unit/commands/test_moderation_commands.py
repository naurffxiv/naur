from datetime import timedelta

import discord
import pytest
from command_helpers import reply, run, sent, target
from pytest_mock.plugin import MockerFixture

from moddingway.commands.ban_commands import create_ban_commands
from moddingway.commands.exile_commands import RouletteDeleteView, create_exile_commands
from moddingway.commands.forum_commands import create_forum_commands
from moddingway.commands.send_message_commands import create_send_message_commands
from moddingway.commands.slowmode_commands import create_slowmode_commands
from moddingway.commands.strikes_command import (
    StrikeDeleteView,
    create_strikes_commands,
)
from moddingway.constants import StrikeSeverity


@pytest.fixture
def mod_bot(bot):
    for create in (
        create_ban_commands,
        create_exile_commands,
        create_forum_commands,
        create_send_message_commands,
        create_slowmode_commands,
        create_strikes_commands,
    ):
        create(bot)
    return bot


# ---- ban ----


@pytest.mark.asyncio
async def test_ban_refuses_to_ban_a_mod(
    mod_bot, make_interaction, mocker: MockerFixture
):
    ban_member = mocker.patch("moddingway.commands.ban_commands.ban_member")
    i = make_interaction("ban")

    await run(mod_bot, "ban", i, user=target(mocker, mod=True), reason="x")

    ban_member.assert_not_called()
    assert reply(i) == "Unable to ban <@12345>: You cannot ban a mod."


@pytest.mark.asyncio
async def test_ban_member(mod_bot, make_interaction, mocker: MockerFixture):
    ban_member = mocker.patch(
        "moddingway.commands.ban_commands.ban_member", return_value=(True, None)
    )
    i = make_interaction("ban")
    user = target(mocker)

    await run(mod_bot, "ban", i, user=user, reason="spam", delete_messages=True)

    i.response.defer.assert_awaited_once_with(ephemeral=True)
    assert ban_member.await_args.args[1:] == (user, "spam", True)
    assert reply(i) == "Successfully banned <@12345>."


@pytest.mark.asyncio
async def test_ban_user_not_in_server_bans_by_id(
    mod_bot, make_interaction, mocker: MockerFixture
):
    ban_user = mocker.patch(
        "moddingway.commands.ban_commands.ban_user",
        return_value=(False, "Failed to ban <@12345>."),
    )
    i = make_interaction("ban")
    i.guild.get_member.return_value = None
    user = target(mocker, member=False)

    await run(mod_bot, "ban", i, user=user, reason="spam")

    assert ban_user.await_args.args[1:] == (i, user, "spam", False)
    assert reply(i) == "Failed to ban <@12345>."


@pytest.mark.asyncio
async def test_ban_user_who_is_in_server_is_checked_as_member(
    mod_bot, make_interaction, mocker: MockerFixture
):
    ban_member = mocker.patch("moddingway.commands.ban_commands.ban_member")
    i = make_interaction("ban")
    i.guild.get_member.return_value = target(mocker, mod=True)

    await run(mod_bot, "ban", i, user=target(mocker, member=False), reason="spam")

    ban_member.assert_not_called()
    assert reply(i) == "Unable to ban <@12345>: You cannot ban a mod."


# ---- exile ----


@pytest.mark.asyncio
async def test_exile(mod_bot, make_interaction, mocker: MockerFixture):
    exile_user = mocker.patch(
        "moddingway.commands.exile_commands.exile_user", return_value=None
    )
    i = make_interaction("exile")
    user = target(mocker)

    await run(mod_bot, "exile", i, user=user, duration="2hour", reason="spam")

    assert exile_user.await_args.args[1:3] == (
        user,
        timedelta(hours=2),
    )
    assert reply(i) == "Successfully exiled <@12345>"


@pytest.mark.parametrize(
    "duration,mod,expected",
    [
        (
            "forever",
            False,
            "Invalid exile duration given, duration should be in the form value between 1 and 99 followed by sec, min, hour, or day. Examples: 1sec, 1min, 1hour, 1day. No action will be taken",
        ),
        ("1day", True, "Unable to exile <@12345>: You cannot exile a mod."),
    ],
)
@pytest.mark.asyncio
async def test_exile_rejections(
    mod_bot, make_interaction, mocker: MockerFixture, duration, mod, expected
):
    exile_user = mocker.patch("moddingway.commands.exile_commands.exile_user")
    i = make_interaction("exile")

    await run(
        mod_bot,
        "exile",
        i,
        user=target(mocker, mod=mod),
        duration=duration,
        reason="spam",
    )

    exile_user.assert_not_called()
    assert sent(i) == expected


@pytest.mark.asyncio
async def test_exile_reports_service_errors(
    mod_bot, make_interaction, mocker: MockerFixture
):
    mocker.patch(
        "moddingway.commands.exile_commands.exile_user",
        return_value="User is not currently verified, no action will be taken",
    )
    i = make_interaction("exile")

    await run(mod_bot, "exile", i, user=target(mocker), duration="1day", reason="spam")

    assert reply(i) == "User is not currently verified, no action will be taken"


@pytest.mark.parametrize(
    "error,expected",
    [
        (None, "Successfully unexiled <@12345>"),
        (
            "User is not currently exiled, no action will be taken",
            "User is not currently exiled, no action will be taken",
        ),
    ],
)
@pytest.mark.asyncio
async def test_unexile(
    mod_bot, make_interaction, mocker: MockerFixture, error, expected
):
    mocker.patch("moddingway.commands.exile_commands.unexile_user", return_value=error)
    i = make_interaction("unexile")

    await run(mod_bot, "unexile", i, user=target(mocker))

    assert reply(i) == expected


@pytest.mark.parametrize(
    "command,patched,kwargs",
    [
        ("view_active_exiles", "get_active_exiles", {}),
        ("delete_exile", "delete_exile_by_id", {"exile_id": 3}),
        ("run_auto_unexile", "autounexile_users", {}),
    ],
)
@pytest.mark.asyncio
async def test_exile_commands_reply_with_service_result(
    mod_bot, make_interaction, mocker: MockerFixture, command, patched, kwargs
):
    mocker.patch(
        f"moddingway.commands.exile_commands.{patched}",
        new=mocker.AsyncMock(return_value="done"),
    )
    i = make_interaction(command)

    await run(mod_bot, command, i, **kwargs)

    assert reply(i) == "done"


@pytest.mark.asyncio
async def test_view_logged_exiles_handles_errors(
    mod_bot, make_interaction, mocker: MockerFixture
):
    mocker.patch(
        "moddingway.commands.exile_commands.get_user_exiles",
        side_effect=RuntimeError("db down"),
    )
    i = make_interaction("view_logged_exiles")

    await run(mod_bot, "view_logged_exiles", i, user=target(mocker))

    assert reply(i) == "An error occurred while fetching user exiles."


@pytest.mark.asyncio
async def test_roulette_asks_for_confirmation(mod_bot, make_interaction):
    i = make_interaction("roulette")

    await run(mod_bot, "roulette", i)

    assert (
        sent(i)
        == "Are you sure you want to do this? You have a 1/6 chance in being exiled for up to 24 hours."
    )
    assert isinstance(
        i.response.send_message.await_args.kwargs["view"], RouletteDeleteView
    )


@pytest.mark.parametrize(
    "safe,mod,expected",
    [
        (True, False, "<@2> has tested their luck and lives another day..."),
        (
            False,
            True,
            "<@2> has tested their luck and has utterly failed! <@2> has been sent into exile for 6 hour(s).",
        ),
    ],
)
@pytest.mark.asyncio
async def test_roulette_outcomes_without_exile(
    make_interaction, mocker: MockerFixture, safe, mod, expected
):
    mocker.patch("moddingway.commands.exile_commands.choice", side_effect=[safe, 6])
    exile_user = mocker.patch("moddingway.commands.exile_commands.exile_user")
    original = make_interaction("roulette")
    spin = make_interaction(user=target(mocker, id=2, mod=mod))

    await RouletteDeleteView(interaction=original).confirm_button_callback.callback(
        spin
    )

    exile_user.assert_not_called()
    assert sent(spin) == expected


@pytest.mark.asyncio
async def test_roulette_loss_exiles_the_player(make_interaction, mocker: MockerFixture):
    mocker.patch("moddingway.commands.exile_commands.choice", side_effect=[False, 6])
    exile_user = mocker.patch(
        "moddingway.commands.exile_commands.exile_user", return_value=None
    )
    player = target(mocker, id=2)
    spin = make_interaction(user=player)

    await RouletteDeleteView(
        interaction=make_interaction("roulette")
    ).confirm_button_callback.callback(spin)

    assert exile_user.await_args.args[1:] == (
        player,
        timedelta(hours=6),
        "lost roulette",
    )
    assert (
        reply(spin)
        == "<@2> has tested their luck and has utterly failed! <@2> has been sent into exile for 6 hour(s)."
    )


@pytest.mark.asyncio
async def test_roulette_cancel(make_interaction):
    cancel = make_interaction()

    await RouletteDeleteView(
        interaction=make_interaction("roulette")
    ).cancel_button_callback.callback(cancel)

    assert reply(cancel) == "Roulette cancelled"


# ---- strikes ----


@pytest.mark.asyncio
async def test_add_strike(mod_bot, make_interaction, mocker: MockerFixture):
    add_strike = mocker.patch("moddingway.services.strike_service.add_strike")
    i = make_interaction("add_strike")
    user = target(mocker)

    await run(
        mod_bot,
        "add_strike",
        i,
        user=user,
        severity=StrikeSeverity.MODERATE,
        reason="spam",
    )

    assert add_strike.await_args.args[1:] == (
        user,
        StrikeSeverity.MODERATE,
        "spam",
        i.user,
    )
    assert reply(i) == "Successfully added strike to <@12345>"


@pytest.mark.asyncio
async def test_add_strike_refuses_mods(
    mod_bot, make_interaction, mocker: MockerFixture
):
    add_strike = mocker.patch("moddingway.services.strike_service.add_strike")
    i = make_interaction("add_strike")

    await run(
        mod_bot,
        "add_strike",
        i,
        user=target(mocker, mod=True),
        severity=StrikeSeverity.MINOR,
        reason="x",
    )

    add_strike.assert_not_called()
    assert (
        sent(i) == "Unable to add strike to <@12345>: You cannot add strike to a mod."
    )


@pytest.mark.asyncio
async def test_view_strikes_and_decrement(
    mod_bot, make_interaction, mocker: MockerFixture
):
    mocker.patch(
        "moddingway.services.strike_service.get_user_strikes", return_value="strikes"
    )
    mocker.patch(
        "moddingway.commands.strikes_command.decrement_strikes",
        new=mocker.AsyncMock(return_value="decremented 2"),
    )

    i = make_interaction("view_strikes")
    await run(mod_bot, "view_strikes", i, user=target(mocker))
    assert reply(i) == "strikes"

    i = make_interaction("run_strike_decrement")
    await run(mod_bot, "run_strike_decrement", i)
    assert reply(i) == "decremented 2"


@pytest.mark.asyncio
async def test_delete_strike_confirm_and_cancel(
    mod_bot, make_interaction, mocker: MockerFixture
):
    delete = mocker.patch(
        "moddingway.services.strike_service.delete_strike",
        return_value="Successfully deleted strike 5",
    )
    i = make_interaction("delete_strike")

    await run(mod_bot, "delete_strike", i, strike_id=5)
    view = i.response.send_message.await_args.kwargs["view"]
    assert isinstance(view, StrikeDeleteView)

    confirm = make_interaction()
    await view.confirm_button_callback.callback(confirm)
    assert delete.await_args.args[1] == view.strike_id
    assert reply(confirm) == "Successfully deleted strike 5"

    cancel = make_interaction()
    await StrikeDeleteView(strike_id=5, interaction=i).cancel_button_callback.callback(
        cancel
    )
    assert reply(cancel) == "Strike deletion cancelled"


# ---- slowmode, send_message, automod ----


@pytest.mark.asyncio
async def test_set_slowmode_defaults_to_the_current_channel(
    mod_bot, make_interaction, mocker: MockerFixture
):
    edit = mocker.patch(
        "moddingway.commands.slowmode_commands.edit_slowmode",
        return_value="Successfully set slowmode",
    )
    i = make_interaction("set_slowmode")

    await run(mod_bot, "set_slowmode", i, interval=30)

    assert edit.await_args.kwargs["channel"] is i.channel
    assert edit.await_args.kwargs["interval"] == 30  # noqa: PLR2004 - the interval passed above
    assert reply(i) == "Successfully set slowmode"


@pytest.mark.asyncio
async def test_set_slowmode_needs_a_text_channel(
    mod_bot, make_interaction, mocker: MockerFixture
):
    edit = mocker.patch("moddingway.commands.slowmode_commands.edit_slowmode")
    i = make_interaction("set_slowmode")
    i.channel = mocker.Mock(spec=discord.Thread)

    await run(mod_bot, "set_slowmode", i, interval=30)

    edit.assert_not_called()
    assert sent(i) == "This command can only be used in a text channel."


@pytest.mark.asyncio
async def test_send_message_and_run_automod(
    mod_bot, make_interaction, mocker: MockerFixture
):
    send = mocker.patch(
        "moddingway.commands.send_message_commands.send_message_to_channel",
        return_value="Successfully sent message to #news",
    )
    mocker.patch(
        "moddingway.commands.forum_commands.autodelete_threads",
        new=mocker.AsyncMock(return_value="removed 3"),
    )
    channel = mocker.Mock(spec=discord.TextChannel, mention="#news")

    i = make_interaction("send_message")
    await run(mod_bot, "send_message", i, channel=channel, message="hi")
    assert send.await_args.args[1:] == (channel, "hi")
    assert reply(i) == "Successfully sent message to #news"

    i = make_interaction("run_automod")
    await run(mod_bot, "run_automod", i)
    assert reply(i) == "removed 3"
