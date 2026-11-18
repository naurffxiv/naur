import logging
from datetime import UTC, datetime, timedelta

from discord.ext import tasks

from moddingway.settings import get_settings
from moddingway.util import (
    create_interaction_embed_context,
    get_log_channel,
)

from .inactivity_purge import (
    ChannelReport,
    PurgeRule,
    describe,
    purge,
    rules_from_setting,
)

settings = get_settings()
logger = logging.getLogger(__name__)


async def _run(bot, name: str, purge_rules: list[PurgeRule]) -> bool:
    logger.info(f"Started {name} worker task.")
    guild = bot.get_guild(settings.guild_id)
    if guild is None:
        logger.error("Guild not found.")
        logger.info(f"Ended {name} worker task with errors.")
        return False

    reports = await purge(purge_rules, guild, datetime.now(UTC))
    await log_reports(bot, reports)
    logger.info(f"Completed {name} worker task.")
    return True


async def log_reports(bot, reports: list[ChannelReport]) -> None:
    log_channel = get_log_channel(bot)
    for report in reports:
        description = describe(report)
        if description is None:
            continue
        try:
            async with create_interaction_embed_context(
                log_channel,
                user=bot.user,
                timestamp=datetime.now(UTC),
                description=description,
            ):
                pass
        except Exception as e:
            logger.error(e, exc_info=e)


# Worker for thread automodding
@tasks.loop(hours=24)
async def autodelete_threads(self):
    forum_rules = rules_from_setting(settings.automod_inactivity, timedelta(days=1))
    if await _run(self, "forum automod", forum_rules):
        return "Forum automod task completed."


# Worker for channel message automodding
@tasks.loop(hours=1)
async def autodelete_posts(self):
    channel_rules = rules_from_setting(
        settings.channel_automod_inactivity, timedelta(minutes=1)
    )
    if await _run(self, "channel automod", channel_rules):
        return "Channel automod task completed."


@autodelete_threads.before_loop
async def before_autodelete_threads():
    logger.info("Forum Automod started, task running every 24 hours.")
