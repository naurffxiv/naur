import runpy
from pathlib import Path

import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import bot as bot_module

MAIN = Path(__file__).resolve().parents[3] / "services" / "moddingway" / "main.py"


def _bot():
    return bot_module.ModdingwayBot(command_prefix="/", intents=discord.Intents.none())


@pytest.mark.asyncio
async def test_setup_hook_registers_everything_and_syncs_to_the_guild(
    mocker: MockerFixture,
):
    bot = _bot()
    sync = mocker.patch.object(bot.tree, "sync", new=mocker.AsyncMock())
    copy = mocker.patch.object(bot.tree, "copy_global_to")

    await bot.setup_hook()

    names = {c.name for c in bot.tree.get_commands()}
    assert {
        "exile",
        "ban",
        "add_strike",
        "warn",
        "draft_announcement",
        "run_automod",
    } <= names
    assert hasattr(bot, "on_member_join")
    guild = copy.call_args.kwargs["guild"]
    assert guild.id == bot_module.settings.guild_id
    sync.assert_awaited_once_with(guild=guild)


@pytest.mark.asyncio
async def test_on_ready_starts_the_workers(monkeypatch):
    started = []
    monkeypatch.setattr(bot_module.workers, "start_tasks", started.append)
    bot = _bot()

    await bot.on_ready()

    assert started == [bot]


def test_main_connects_to_the_database_then_runs_the_bot(mocker: MockerFixture):
    bot_class = mocker.patch("moddingway.bot.ModdingwayBot")
    db_class = mocker.patch("moddingway.database.DatabaseConnection")
    mocker.patch("logging.basicConfig")

    runpy.run_path(str(MAIN), run_name="__main__")

    intents = bot_class.call_args.kwargs["intents"]
    assert intents.members and intents.message_content
    db = db_class.return_value
    assert db.method_calls[:2] == [mocker.call.connect(), mocker.call.create_tables()]
    bot_class.return_value.run.assert_called_once()
