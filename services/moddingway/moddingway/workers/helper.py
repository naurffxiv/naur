import asyncio
import logging
from datetime import UTC, datetime, timedelta

import discord
from discord.utils import snowflake_time

from moddingway.settings import get_settings
from moddingway.util import (
    create_interaction_embed_context,
    get_log_channel,
)

settings = get_settings()
logger = logging.getLogger(__name__)


def create_autounexile_embed(
    self,
    user: discord.Member | None,
    discord_id: int,
    exile_id: str,
    end_timestamp: str,
):
    return create_interaction_embed_context(
        get_log_channel(self),
        user=user,
        timestamp=end_timestamp,
        description=f"<@{discord_id}>'s exile has timed out",
        footer=f"Exile ID: {exile_id}",
    )


def create_automod_embed(self, channel_id, num_removed, num_error, timestamp: datetime):
    return create_interaction_embed_context(
        get_log_channel(self),
        user=self.user,
        timestamp=timestamp,
        description=f"Successfully removed {num_removed} inactive thread(s) from <#{channel_id}>.\n{num_error} inactive thread(s) failed to be removed.",
    )


def create_channel_automod_embed(
    self, channel_id, num_removed, num_error, timestamp: datetime
):
    return create_interaction_embed_context(
        get_log_channel(self),
        user=self.user,
        timestamp=timestamp,
        description=f"Successfully removed {num_removed} old message(s) from <#{channel_id}>.\n{num_error} old messages(s) failed to be removed.",
    )


async def automod_thread(
    thread: discord.Thread,
    duration: int,
    num_removed: int,
    num_errors: int,
):
    if thread.flags.pinned:
        # skip the for loop if the thread is pinned
        return num_removed, num_errors

    should_delete = _should_delete_thread(thread, duration)

    if not should_delete:
        return num_removed, num_errors

    # Execute deletion
    try:
        await thread.delete()
        logger.info(f"Thread {thread.id} has been deleted successfully")
        return num_removed + 1, num_errors
    except Exception as e:
        logger.error(f"Unexpected error for thread {thread.id}: {e}", exc_info=e)
        return num_removed, num_errors + 1


def _should_delete_thread(
    thread: discord.Thread,
    duration: int,
) -> bool:
    now = datetime.now(UTC)

    # Standard duration check (fallback or if not specific user logic)
    if thread.last_message_id:
        time_since = now - snowflake_time(thread.last_message_id)
        if time_since >= timedelta(days=duration):
            return True

    return False


async def automod_channel(messages, duration: int, channel: str):
    num_removed = 0
    num_errors = 0

    async for message in messages:
        # Skip any pinned messages
        if message.pinned:
            continue

        # Delete message if older than 150 minutes
        now = datetime.now(UTC)
        time_since = now - snowflake_time(message.id)
        if time_since > timedelta(minutes=duration):
            try:
                await message.delete()
                logger.info(
                    f"Message {message.id} has been deleted successfully from #{channel}"
                )
                num_removed += 1
                await asyncio.sleep(1)
            except Exception as e:
                logger.error(
                    f"Unexpected error for message {message.id}: {e}", exc_info=e
                )
                num_errors += 1

    return num_removed, num_errors
