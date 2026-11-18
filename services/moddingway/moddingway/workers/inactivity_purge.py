"""Inactivity purge: remove old forum threads and old channel messages.

The channel's type decides what is purged: forum channels lose inactive
threads, every other channel with a message history (text, announcement,
voice chat...) loses old messages. Pinned threads and messages are
always kept. Each channel pass is bounded by a timeout so one hung channel
cannot stall the others.
"""

import asyncio
import logging
from collections.abc import Iterable, Mapping
from dataclasses import dataclass
from datetime import datetime, timedelta
from enum import Enum
from typing import Any, Protocol

import discord
from discord.utils import snowflake_time

logger = logging.getLogger(__name__)

CHANNEL_PASS_TIMEOUT = timedelta(minutes=30)
MESSAGE_DELETE_PAUSE = timedelta(seconds=1)


class PurgeKind(Enum):
    THREADS = "threads"
    MESSAGES = "messages"


_THREAD_CHANNEL_TYPES = {discord.ChannelType.forum, discord.ChannelType.media}


@dataclass(frozen=True)
class PurgeRule:
    channel_id: int
    max_age: timedelta


def rules_from_setting(limits: Mapping[int, int], unit: timedelta) -> list[PurgeRule]:
    """Rules from a {channel id: limit} setting, where each limit counts `unit`s."""
    return [PurgeRule(cid, limit * unit) for cid, limit in limits.items()]


class ChannelSource(Protocol):
    """Where channels come from; a discord.Guild satisfies this."""

    def get_channel(self, channel_id: int, /) -> Any: ...


@dataclass
class ChannelReport:
    channel_id: int
    kind: PurgeKind | None = None
    removed: int = 0
    errors: int = 0
    failure: Exception | None = None


async def purge(
    rules: Iterable[PurgeRule],
    source: ChannelSource,
    now: datetime,
    *,
    timeout: timedelta = CHANNEL_PASS_TIMEOUT,
    message_delete_pause: timedelta = MESSAGE_DELETE_PAUSE,
) -> list[ChannelReport]:
    reports = []
    for rule in rules:
        report = ChannelReport(rule.channel_id)
        reports.append(report)
        try:
            async with asyncio.timeout(timeout.total_seconds()):
                await _purge_channel(
                    rule, source, now, report, message_delete_pause.total_seconds()
                )
        except Exception as e:
            report.failure = e
            logger.error(e, exc_info=e)
    return reports


async def _purge_channel(
    rule: PurgeRule,
    source: ChannelSource,
    now: datetime,
    report: ChannelReport,
    pause_seconds: float,
) -> None:
    channel = source.get_channel(rule.channel_id)
    if channel is None:
        logger.error(f"Channel {rule.channel_id} not found.")
        return

    kind = report.kind = _kind_of(channel)

    cutoff = now - rule.max_age
    if kind is PurgeKind.THREADS:
        async for thread in channel.archived_threads(limit=None):
            await _purge_thread(thread, cutoff, report)
        for thread in channel.threads:
            await _purge_thread(thread, cutoff, report)
    else:
        # Discord filters by message ID, so newer messages are never fetched.
        async for message in channel.history(limit=None, before=cutoff):
            if message.pinned or snowflake_time(message.id) > cutoff:
                continue
            if await _delete(message, "Message", report):
                await asyncio.sleep(pause_seconds)

    if report.removed or report.errors:
        logger.info(
            f"Removed a total of {report.removed} {kind.value} from channel "
            f"{rule.channel_id}. {report.errors} failed removals."
        )
    else:
        logger.info(
            f"No {kind.value} were marked for deletion in channel {rule.channel_id}."
        )


def _kind_of(channel) -> PurgeKind:
    if channel.type in _THREAD_CHANNEL_TYPES:
        return PurgeKind.THREADS
    if hasattr(channel, "history"):
        return PurgeKind.MESSAGES
    raise TypeError(f"Unsupported channel type for purge: {channel.type}")


async def _purge_thread(
    thread: discord.Thread, cutoff: datetime, report: ChannelReport
) -> None:
    if thread.flags.pinned or thread.last_message_id is None:
        return
    if snowflake_time(thread.last_message_id) > cutoff:
        return
    await _delete(thread, "Thread", report)


async def _delete(
    item: discord.Thread | discord.Message, label: str, report: ChannelReport
) -> bool:
    """Deletes a thread or message and counts the outcome; True when it was deleted."""
    try:
        await item.delete()
    except Exception as e:
        logger.error(f"Unexpected error for {label.lower()} {item.id}: {e}", exc_info=e)
        report.errors += 1
        return False
    logger.info(f"{label} {item.id} has been deleted successfully")
    report.removed += 1
    return True


def describe(report: ChannelReport) -> str | None:
    """The logging-channel embed description for a report, or None to stay quiet."""
    channel = f"<#{report.channel_id}>"
    if report.failure is not None:
        reason = str(report.failure)
        if report.kind is not PurgeKind.MESSAGES:
            # The thread pass has always named the exception when it has no message.
            reason = reason or type(report.failure).__name__
        return f"Automod task failed to process channel {channel}: {reason}"
    if report.kind is PurgeKind.THREADS and (report.removed or report.errors):
        return (
            f"Successfully removed {report.removed} inactive thread(s) from {channel}."
            f"\n{report.errors} inactive thread(s) failed to be removed."
        )
    if report.kind is PurgeKind.MESSAGES and report.errors:
        return (
            f"Successfully removed {report.removed} old message(s) from {channel}."
            f"\n{report.errors} old messages(s) failed to be removed."
        )
    return None
