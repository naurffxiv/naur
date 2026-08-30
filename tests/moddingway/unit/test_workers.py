import asyncio

import discord
import pytest
from discord.ext import tasks
from pytest_mock.plugin import MockerFixture

from moddingway import workers

TICK = 0.05


@pytest.fixture
async def create_worker():
    created = []

    def __create_worker(error: Exception | None = None):
        calls = []

        @tasks.loop(seconds=30)
        async def worker(bot):
            calls.append(bot)
            if error is not None:
                raise error

        created.append(worker)
        return worker, calls

    yield __create_worker

    for worker in created:
        worker.cancel()
    await asyncio.sleep(0)


@pytest.fixture
def discord_outage(mocker: MockerFixture):
    response = mocker.Mock(status=503, reason="Service Unavailable")
    return discord.DiscordServerError(response, "503 Service Unavailable")


async def test_start_tasks_starts_every_worker(monkeypatch, create_worker):
    first, first_calls = create_worker()
    second, second_calls = create_worker()
    monkeypatch.setattr(workers, "WORKERS", (first, second))
    bot = object()

    workers.start_tasks(bot)
    await asyncio.sleep(TICK)

    assert first.is_running()
    assert second.is_running()
    assert first_calls == [bot]
    assert second_calls == [bot]


async def test_start_tasks_starts_remaining_workers_when_one_is_running(
    monkeypatch, create_worker
):
    running, _ = create_worker()
    pending, pending_calls = create_worker()
    monkeypatch.setattr(workers, "WORKERS", (running, pending))
    bot = object()
    running.start(bot)
    await asyncio.sleep(TICK)

    workers.start_tasks(bot)
    await asyncio.sleep(TICK)

    assert pending.is_running()
    assert pending_calls == [bot]


async def test_start_tasks_restarts_a_worker_that_stopped(monkeypatch, create_worker):
    worker, calls = create_worker()
    monkeypatch.setattr(workers, "WORKERS", (worker,))
    bot = object()
    workers.start_tasks(bot)
    await asyncio.sleep(TICK)
    worker.cancel()
    await asyncio.sleep(TICK)
    assert not worker.is_running()

    workers.start_tasks(bot)
    await asyncio.sleep(TICK)

    assert worker.is_running()
    assert calls == [bot, bot]


async def test_worker_keeps_running_after_a_discord_outage(
    monkeypatch, create_worker, discord_outage
):
    worker, calls = create_worker(error=discord_outage)
    monkeypatch.setattr(workers, "WORKERS", (worker,))

    workers.start_tasks(object())
    await asyncio.sleep(TICK)

    assert calls
    assert worker.is_running()
    assert not worker.failed()


async def test_worker_keeps_running_after_an_unexpected_error(
    monkeypatch, create_worker
):
    error = TypeError("object TextChannel can't be used in await expression")
    worker, calls = create_worker(error=error)
    monkeypatch.setattr(workers, "WORKERS", (worker,))

    workers.start_tasks(object())
    await asyncio.sleep(TICK)

    assert calls
    assert worker.is_running()
    assert not worker.failed()
