import asyncio
import contextlib
from datetime import timedelta

import pytest
from pytest_mock.plugin import MockerFixture

from moddingway.workers import forum_automod

HANGING_CHANNEL_ID = 1
HEALTHY_CHANNEL_ID = 2


class HangingThreads:
    def __aiter__(self):
        return self

    async def __anext__(self):
        await asyncio.Event().wait()


class NoThreads:
    def __aiter__(self):
        return self

    async def __anext__(self):
        raise StopAsyncIteration


@pytest.fixture
def reported(monkeypatch):
    descriptions = []

    @contextlib.asynccontextmanager
    async def fake_interaction_embed(channel, **kwargs):
        descriptions.append(kwargs["description"])
        yield

    @contextlib.asynccontextmanager
    async def fake_automod_embed(*args):
        yield

    monkeypatch.setattr(
        forum_automod, "create_interaction_embed_context", fake_interaction_embed
    )
    monkeypatch.setattr(forum_automod, "create_automod_embed", fake_automod_embed)
    monkeypatch.setattr(forum_automod, "get_log_channel", lambda bot: None)
    return descriptions


@pytest.fixture
def forum_bot(mocker: MockerFixture, monkeypatch):
    healthy_thread = mocker.Mock()
    channels = {
        HANGING_CHANNEL_ID: mocker.Mock(
            archived_threads=lambda limit: HangingThreads(), threads=[]
        ),
        HEALTHY_CHANNEL_ID: mocker.Mock(
            archived_threads=lambda limit: NoThreads(), threads=[healthy_thread]
        ),
    }
    guild = mocker.Mock()
    guild.get_channel.side_effect = lambda channel_id: channels.get(
        channel_id, mocker.Mock()
    )
    bot = mocker.Mock()
    bot.get_guild.return_value = guild

    monkeypatch.setattr(
        forum_automod.settings,
        "automod_inactivity",
        {HANGING_CHANNEL_ID: 30, HEALTHY_CHANNEL_ID: 30},
    )
    automod_thread = mocker.AsyncMock(
        side_effect=lambda thread, duration, removed, errors: (removed + 1, errors)
    )
    monkeypatch.setattr(forum_automod, "automod_thread", automod_thread)
    return bot, automod_thread, healthy_thread


async def test_thread_worker_moves_past_a_channel_that_hangs(
    monkeypatch, forum_bot, reported
):
    bot, automod_thread, healthy_thread = forum_bot
    monkeypatch.setattr(
        forum_automod, "CHANNEL_PASS_TIMEOUT", timedelta(milliseconds=50)
    )

    result = await asyncio.wait_for(forum_automod.autodelete_threads.coro(bot), 1)

    assert result == "Forum automod task completed."
    assert automod_thread.await_args.args[0] is healthy_thread
    assert len(reported) == 1
    assert f"<#{HANGING_CHANNEL_ID}>" in reported[0]
    assert "TimeoutError" in reported[0]
