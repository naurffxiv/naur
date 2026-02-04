import logging

import discord

from moddingway.constants import MAX_SLOWMODE_INTERVAL
from moddingway.util import log_info_and_add_field

logger = logging.getLogger(__name__)


async def edit_slowmode(
    logging_embed: discord.Embed,
    interval: int,
    channel: discord.TextChannel,
) -> str:
    if interval < 0 or interval > MAX_SLOWMODE_INTERVAL:
        error_message = (
            f"Interval must be between 0 and {MAX_SLOWMODE_INTERVAL} seconds"
        )
        log_info_and_add_field(logging_embed, logger, "Error", error_message)
        return error_message

    # check if channel is already set to the desired entry
    if channel.slowmode_delay == interval:
        if interval == 0:
            error_message = f"Slowmode is already off in {channel.mention}"
        else:
            error_message = (
                f"Slowmode is already set to {interval} seconds in {channel.mention}"
            )
        log_info_and_add_field(logging_embed, logger, "Error", error_message)
        return error_message

    # set slowmode
    await channel.edit(slowmode_delay=interval)
    result = ""
    if interval == 0:
        result = f"Successfully turned off slowmode in {channel.mention}"
    else:
        result = f"Successfully set slowmode to {interval} seconds in {channel.mention}"
    log_info_and_add_field(logging_embed, logger, "Result", result)
    return result
