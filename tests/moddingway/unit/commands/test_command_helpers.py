import datetime

import discord
import pytest
from fakes import embed_fields
from pytest_mock.plugin import MockerFixture

from moddingway.commands.helper import (
    create_bot_errors,
    create_deferred_response_context,
    create_logging_embed,
    create_response_context,
)


@pytest.mark.asyncio
async def test_response_context_replaces_placeholder_with_final_message(
    make_interaction,
):
    i = make_interaction()

    async with create_response_context(i) as response:
        response.set_string("first")
        response.append_string("second")

    i.response.send_message.assert_awaited_once_with("Processing...", ephemeral=True)
    i.original.edit.assert_awaited_once_with(content="first\nsecond")


@pytest.mark.asyncio
async def test_response_context_without_message_says_so(make_interaction):
    i = make_interaction()

    async with create_response_context(i, sendEphemeral=False):
        pass

    i.response.send_message.assert_awaited_once_with("Processing...", ephemeral=False)
    i.original.edit.assert_awaited_once_with(
        content="Command finished without a response."
    )


@pytest.mark.asyncio
async def test_response_context_reports_errors_and_reraises(make_interaction):
    i = make_interaction()

    with pytest.raises(RuntimeError):
        async with create_response_context(i) as response:
            response.set_string("partial")
            raise RuntimeError("boom")

    i.original.edit.assert_awaited_once_with(
        content="partial\nAn unexpected error has occurred"
    )


@pytest.mark.asyncio
async def test_deferred_response_context_edits_the_deferred_message(make_interaction):
    i = make_interaction()

    async with create_deferred_response_context(i) as response:
        response.set_string("done")

    assert [c.kwargs["content"] for c in i.original.edit.await_args_list] == [
        "Processing...",
        "done",
    ]
    i.response.send_message.assert_not_awaited()


@pytest.mark.asyncio
async def test_logging_embed_formats_arguments_and_posts_to_log_channel(
    make_interaction, mocker: MockerFixture
):
    i = make_interaction("exile")
    when = datetime.datetime(2026, 1, 2, tzinfo=datetime.UTC)

    async with create_logging_embed(
        i, reason="spam", end_time=when, exile_id=3
    ) as embed:
        embed.add_field(name="Result", value="ok")

    posted = i.log_channel.send.await_args.kwargs["embed"]
    assert posted is embed
    assert posted.description == "Used `exile` command in #mod-chat"
    assert posted.author.name == "mod"
    assert embed_fields(posted) == [
        ("Action", "/exile"),
        ("Reason", "spam"),
        ("End Time", f"<t:{int(when.timestamp())}:R>"),
        ("Exile Id", "3"),
        ("Result", "ok"),
    ]


@pytest.mark.asyncio
async def test_logging_embed_for_ui_interactions(make_interaction):
    i = make_interaction()
    i.command = None

    async with create_logging_embed(i, custom_response="Lost roulette") as embed:
        pass
    assert embed.description == "Lost roulette"

    async with create_logging_embed(i) as embed:
        pass
    assert embed.description == "Command was run via a UI"


@pytest.mark.asyncio
async def test_logging_embed_records_errors(make_interaction):
    i = make_interaction("exile")

    with pytest.raises(ValueError):
        async with create_logging_embed(i):
            raise ValueError("bad input")

    assert ("Error", "bad input") in embed_fields(
        i.log_channel.send.await_args.kwargs["embed"]
    )


@pytest.mark.parametrize(
    "error,expected",
    [
        (
            discord.app_commands.MissingRole("Mod"),
            "You are missing the required 'Mod' role to use this command.",
        ),
        (
            discord.app_commands.CheckFailure("not a mod"),
            "You do not meet the requirements: not a mod",
        ),
        (
            RuntimeError("boom"),
            "An unexpected error occurred. Please contact an admin.",
        ),
    ],
)
@pytest.mark.asyncio
async def test_bot_error_handler(bot, make_interaction, error, expected):
    create_bot_errors(bot)
    i = make_interaction()

    await bot.tree.on_error(i, error)

    i.response.send_message.assert_awaited_once_with(expected, ephemeral=True)


@pytest.mark.asyncio
async def test_bot_error_handler_reports_cooldowns(
    bot, make_interaction, mocker: MockerFixture
):
    mocker.patch("moddingway.commands.helper.time.time", return_value=1000)
    create_bot_errors(bot)
    i = make_interaction()

    await bot.tree.on_error(
        i,
        discord.app_commands.CommandOnCooldown(
            discord.app_commands.Cooldown(1, 60), 59.5
        ),
    )

    i.response.send_message.assert_awaited_once_with(
        "This command is on cooldown. Try again <t:1059:R>.", ephemeral=True
    )


@pytest.mark.asyncio
async def test_bot_error_handler_tolerates_already_responded(bot, make_interaction):
    create_bot_errors(bot)
    i = make_interaction()
    i.response.send_message.side_effect = discord.InteractionResponded(i)

    await bot.tree.on_error(i, RuntimeError("boom"))
