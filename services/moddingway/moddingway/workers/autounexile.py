import logging
from datetime import UTC, datetime

from discord.ext import tasks

from moddingway.services import exiles
from moddingway.services.exile_service import log_exile_result
from moddingway.settings import get_settings

from .helper import create_autounexile_embed

settings = get_settings()
logger = logging.getLogger(__name__)


@tasks.loop(minutes=1.0)
async def autounexile_users(self):
    guild = self.get_guild(settings.guild_id)
    if guild is None:
        logger.error(f"Guild {settings.guild_id} not found, skipping auto unexile.")
        return

    try:
        results = await exiles.release_due(guild, datetime.now(UTC))
    except Exception:
        logger.exception("Failed to get pending exiles.")
        logger.info("Ended auto unexile worker task with errors.")
        return

    for result in results:
        if result.outcome == exiles.ExileOutcome.FAILED:
            logger.info(
                f"Auto Unexile failed for exile {result.exile_id}, user "
                f"{result.discord_id}; it stays active and will be retried."
            )
        try:
            async with create_autounexile_embed(
                self,
                result.member,
                result.discord_id,
                str(result.exile_id),
                result.end_timestamp,
            ) as autounexile_embed:
                log_exile_result(autounexile_embed, result)
        except Exception:
            logger.exception(
                f"Failed to log auto unexile of exile {result.exile_id} to Discord."
            )

    return "Auto Unexile task completed."


@autounexile_users.before_loop
async def before_autounexile_users():
    logger.info("Auto Unexile started, task running every 1 minute.")
