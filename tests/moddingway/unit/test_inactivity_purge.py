import asyncio
from datetime import UTC, datetime, timedelta
from types import SimpleNamespace

import discord
import pytest
from discord.utils import snowflake_time, time_snowflake
from pytest_mock.plugin import MockerFixture

from moddingway.workers import forum_automod
from moddingway.workers.inactivity_purge import (
    ChannelReport,
    PurgeKind,
    PurgeRule,
    describe,
    purge,
    rules_from_setting,
)

NOW = datetime(2026, 1, 1, 12, 0, tzinfo=UTC)
FORUM_ID = 10
TEXT_ID = 20
OTHER_ID = 30


class FakeDeletable:
    def __init__(self, *, fails):
        self.fails = fails
        self.deleted = False

    async def delete(self):
        if self.fails:
            raise RuntimeError("delete failed")
        self.deleted = True


class FakeThread(FakeDeletable):
    def __init__(self, last_active, *, pinned=False, fails=False):
        super().__init__(fails=fails)
        self.id = id(self)
        self.last_message_id = (
            None if last_active is None else time_snowflake(last_active)
        )
        self.flags = SimpleNamespace(pinned=pinned)


class FakeMessage(FakeDeletable):
    def __init__(self, created, *, pinned=False, fails=False):
        super().__init__(fails=fails)
        self.id = time_snowflake(created)
        self.pinned = pinned


async def _stream(items, hang):
    for item in items:
        yield item
    if hang:
        await asyncio.Event().wait()


class FakeForum:
    type = discord.ChannelType.forum

    def __init__(self, active=(), archived=(), *, hang=False, broken=False):
        self.threads = list(active)
        self.archived = list(archived)
        self.hang = hang
        self.broken = broken

    def archived_threads(self, limit):
        assert limit is None
        if self.broken:
            raise RuntimeError("Missing Access")
        return _stream(self.archived, self.hang)


class FakeTextChannel:
    type = discord.ChannelType.text

    def __init__(self, messages=(), *, hang=False):
        self.messages = list(messages)
        self.hang = hang

    def history(self, limit, before):
        assert limit is None
        older = [m for m in self.messages if snowflake_time(m.id) < before]
        return _stream(older, self.hang)


class FakeGuild:
    def __init__(self, channels):
        self.channels = channels

    def get_channel(self, channel_id, /):
        return self.channels.get(channel_id)


async def run_purge(channels, rules, **kwargs):
    kwargs.setdefault("message_delete_pause", timedelta(0))
    return await purge(rules, FakeGuild(channels), NOW, **kwargs)


def ago(**kwargs):
    return NOW - timedelta(**kwargs)


# --- threads ---


async def test_threads_older_than_the_limit_are_removed_and_newer_are_kept():
    old_active = FakeThread(ago(days=8))
    new_active = FakeThread(ago(days=6))
    old_archived = FakeThread(ago(days=30))
    new_archived = FakeThread(ago(hours=1))
    forum = FakeForum([old_active, new_active], [old_archived, new_archived])

    [report] = await run_purge(
        {FORUM_ID: forum}, [PurgeRule(FORUM_ID, timedelta(days=7))]
    )

    assert old_active.deleted and old_archived.deleted
    assert not new_active.deleted and not new_archived.deleted
    assert report == ChannelReport(FORUM_ID, kind=PurgeKind.THREADS, removed=2)


async def test_pinned_and_empty_threads_are_kept():
    pinned = FakeThread(ago(days=90), pinned=True)
    empty = FakeThread(None)
    forum = FakeForum([pinned], [empty])

    [report] = await run_purge(
        {FORUM_ID: forum}, [PurgeRule(FORUM_ID, timedelta(days=7))]
    )

    assert not pinned.deleted and not empty.deleted
    assert report.removed == 0 and report.errors == 0


async def test_failed_thread_deletions_are_counted_and_the_pass_continues():
    failing = FakeThread(ago(days=8), fails=True)
    later = FakeThread(ago(days=8))
    forum = FakeForum([failing, later])

    [report] = await run_purge(
        {FORUM_ID: forum}, [PurgeRule(FORUM_ID, timedelta(days=7))]
    )

    assert later.deleted
    assert (report.removed, report.errors) == (1, 1)


# --- messages ---


async def test_messages_older_than_the_limit_are_removed_and_newer_are_kept():
    old = FakeMessage(ago(minutes=200))
    new = FakeMessage(ago(minutes=100))
    channel = FakeTextChannel([new, old])

    [report] = await run_purge(
        {TEXT_ID: channel}, [PurgeRule(TEXT_ID, timedelta(minutes=150))]
    )

    assert old.deleted and not new.deleted
    assert report == ChannelReport(TEXT_ID, kind=PurgeKind.MESSAGES, removed=1)


async def test_pinned_messages_are_kept():
    pinned = FakeMessage(ago(days=365), pinned=True)
    channel = FakeTextChannel([pinned])

    [report] = await run_purge(
        {TEXT_ID: channel}, [PurgeRule(TEXT_ID, timedelta(minutes=150))]
    )

    assert not pinned.deleted
    assert report.removed == 0


async def test_failed_message_deletions_are_counted_and_the_pass_continues():
    failing = FakeMessage(ago(minutes=200), fails=True)
    later = FakeMessage(ago(minutes=300))
    channel = FakeTextChannel([failing, later])

    [report] = await run_purge(
        {TEXT_ID: channel}, [PurgeRule(TEXT_ID, timedelta(minutes=150))]
    )

    assert later.deleted
    assert (report.removed, report.errors) == (1, 1)


async def test_pauses_between_message_deletions(mocker: MockerFixture):
    sleep = mocker.patch(
        "moddingway.workers.inactivity_purge.asyncio.sleep", mocker.AsyncMock()
    )
    channel = FakeTextChannel([FakeMessage(ago(days=1)), FakeMessage(ago(days=1))])

    await run_purge(
        {TEXT_ID: channel},
        [PurgeRule(TEXT_ID, timedelta(minutes=150))],
        message_delete_pause=timedelta(seconds=1),
    )

    assert sleep.await_args_list == [mocker.call(1.0), mocker.call(1.0)]


# --- channel isolation ---


@pytest.mark.parametrize(
    "hung_channel",
    [
        FakeForum([], [FakeThread(ago(days=8))], hang=True),
        FakeTextChannel([FakeMessage(ago(days=8))], hang=True),
    ],
    ids=["forum", "text"],
)
async def test_a_hung_channel_times_out_and_the_next_channel_is_purged(hung_channel):
    healthy = FakeThread(ago(days=8))
    rules = [
        PurgeRule(OTHER_ID, timedelta(days=7)),
        PurgeRule(FORUM_ID, timedelta(days=7)),
    ]

    hung, ok = await asyncio.wait_for(
        run_purge(
            {OTHER_ID: hung_channel, FORUM_ID: FakeForum([healthy])},
            rules,
            timeout=timedelta(milliseconds=50),
        ),
        1,
    )

    assert isinstance(hung.failure, TimeoutError)
    assert hung.removed == 1
    assert healthy.deleted
    assert ok == ChannelReport(FORUM_ID, kind=PurgeKind.THREADS, removed=1)


async def test_a_failing_channel_does_not_stop_the_next_channel():
    healthy = FakeMessage(ago(days=1))
    rules = [
        PurgeRule(FORUM_ID, timedelta(days=7)),
        PurgeRule(TEXT_ID, timedelta(minutes=150)),
    ]

    broken, ok = await run_purge(
        {FORUM_ID: FakeForum(broken=True), TEXT_ID: FakeTextChannel([healthy])},
        rules,
    )

    assert str(broken.failure) == "Missing Access"
    assert healthy.deleted
    assert ok.removed == 1


async def test_a_missing_channel_is_reported_and_skipped():
    healthy = FakeThread(ago(days=8))
    rules = [
        PurgeRule(OTHER_ID, timedelta(days=7)),
        PurgeRule(FORUM_ID, timedelta(days=7)),
    ]

    missing, ok = await run_purge({FORUM_ID: FakeForum([healthy])}, rules)

    assert missing == ChannelReport(OTHER_ID)
    assert ok.removed == 1


@pytest.mark.parametrize(
    "channel_type",
    [discord.ChannelType.news, discord.ChannelType.voice],
    ids=["announcement", "voice-chat"],
)
async def test_any_channel_with_history_loses_old_messages(channel_type):
    old = FakeMessage(ago(minutes=200))
    channel = FakeTextChannel([old])
    channel.type = channel_type

    [report] = await run_purge(
        {TEXT_ID: channel}, [PurgeRule(TEXT_ID, timedelta(minutes=150))]
    )

    assert report.kind is PurgeKind.MESSAGES
    assert old.deleted


async def test_a_channel_without_history_is_a_failure():
    category = SimpleNamespace(type=discord.ChannelType.category)

    [report] = await run_purge(
        {OTHER_ID: category}, [PurgeRule(OTHER_ID, timedelta(days=1))]
    )

    assert isinstance(report.failure, TypeError)


def test_rules_come_from_day_and_minute_settings():
    assert rules_from_setting({1: 7}, timedelta(days=1)) == [
        PurgeRule(1, timedelta(days=7))
    ]
    assert rules_from_setting({2: 150}, timedelta(minutes=1)) == [
        PurgeRule(2, timedelta(minutes=150))
    ]


# --- logging-channel embeds ---


@pytest.mark.parametrize(
    "report,expected",
    [
        (
            ChannelReport(FORUM_ID, kind=PurgeKind.THREADS, removed=3, errors=1),
            f"Successfully removed 3 inactive thread(s) from <#{FORUM_ID}>.\n"
            "1 inactive thread(s) failed to be removed.",
        ),
        (
            ChannelReport(FORUM_ID, kind=PurgeKind.THREADS, removed=2),
            f"Successfully removed 2 inactive thread(s) from <#{FORUM_ID}>.\n"
            "0 inactive thread(s) failed to be removed.",
        ),
        (ChannelReport(FORUM_ID, kind=PurgeKind.THREADS), None),
        (ChannelReport(TEXT_ID, kind=PurgeKind.MESSAGES, removed=5), None),
        (
            ChannelReport(TEXT_ID, kind=PurgeKind.MESSAGES, removed=5, errors=2),
            f"Successfully removed 5 old message(s) from <#{TEXT_ID}>.\n"
            "2 old messages(s) failed to be removed.",
        ),
        (ChannelReport(OTHER_ID), None),
        (
            ChannelReport(OTHER_ID, failure=RuntimeError("Missing Access")),
            f"Automod task failed to process channel <#{OTHER_ID}>: Missing Access",
        ),
        (
            ChannelReport(
                FORUM_ID, kind=PurgeKind.THREADS, removed=4, failure=TimeoutError()
            ),
            f"Automod task failed to process channel <#{FORUM_ID}>: TimeoutError",
        ),
        (
            ChannelReport(TEXT_ID, kind=PurgeKind.MESSAGES, failure=TimeoutError()),
            f"Automod task failed to process channel <#{TEXT_ID}>: ",
        ),
    ],
)
def test_report_embed_descriptions_match_the_logging_channel_format(report, expected):
    assert describe(report) == expected


# --- workers ---


@pytest.fixture
def log_channel(mocker: MockerFixture):
    return mocker.Mock(spec=discord.TextChannel, send=mocker.AsyncMock())


@pytest.fixture
def make_bot(mocker: MockerFixture, log_channel):
    def __make_bot(guild):
        return mocker.Mock(
            get_guild=mocker.Mock(return_value=guild),
            get_channel=mocker.Mock(return_value=log_channel),
        )

    return __make_bot


def sent_descriptions(log_channel):
    return [c.kwargs["embed"].description for c in log_channel.send.await_args_list]


async def test_thread_worker_purges_and_logs_the_report(
    monkeypatch, make_bot, log_channel
):
    monkeypatch.setattr(forum_automod.settings, "automod_inactivity", {FORUM_ID: 7})
    old = FakeThread(datetime.now(UTC) - timedelta(days=8))
    bot = make_bot(FakeGuild({FORUM_ID: FakeForum([old])}))

    result = await forum_automod.autodelete_threads.coro(bot)

    assert result == "Forum automod task completed."
    assert old.deleted
    assert sent_descriptions(log_channel) == [
        f"Successfully removed 1 inactive thread(s) from <#{FORUM_ID}>.\n"
        "0 inactive thread(s) failed to be removed."
    ]


async def test_post_worker_purges_and_logs_only_failures(
    monkeypatch, mocker: MockerFixture, make_bot, log_channel
):
    monkeypatch.setattr(
        forum_automod.settings,
        "channel_automod_inactivity",
        {TEXT_ID: 150, OTHER_ID: 150},
    )
    old = FakeMessage(datetime.now(UTC) - timedelta(minutes=200))
    channels = {TEXT_ID: FakeTextChannel([old]), OTHER_ID: FakeForum(broken=True)}
    bot = make_bot(FakeGuild(channels))
    mocker.patch(
        "moddingway.workers.inactivity_purge.asyncio.sleep", mocker.AsyncMock()
    )

    result = await forum_automod.autodelete_posts.coro(bot)

    assert result == "Channel automod task completed."
    assert old.deleted
    assert sent_descriptions(log_channel) == [
        f"Automod task failed to process channel <#{OTHER_ID}>: Missing Access"
    ]


@pytest.mark.parametrize(
    "worker", [forum_automod.autodelete_threads, forum_automod.autodelete_posts]
)
async def test_workers_stop_when_the_guild_is_missing(worker, make_bot, log_channel):
    assert await worker.coro(make_bot(None)) is None
    log_channel.send.assert_not_awaited()


async def test_one_failed_embed_does_not_stop_the_rest(make_bot, log_channel):
    log_channel.send.side_effect = [RuntimeError("send failed"), None]
    reports = [
        ChannelReport(1, failure=RuntimeError("a")),
        ChannelReport(2, failure=RuntimeError("b")),
    ]

    await forum_automod.log_reports(make_bot(None), reports)

    assert sent_descriptions(log_channel) == [
        "Automod task failed to process channel <#1>: a",
        "Automod task failed to process channel <#2>: b",
    ]
