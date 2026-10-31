from moddingway import bot as bot_module


async def test_bot_restarts_workers_when_the_gateway_resumes(monkeypatch):
    started = []
    monkeypatch.setattr(bot_module.workers, "start_tasks", started.append)
    bot = object.__new__(bot_module.ModdingwayBot)

    await bot.on_resumed()

    assert started == [bot]
