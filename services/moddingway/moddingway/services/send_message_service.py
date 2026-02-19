import logging

import discord

from moddingway.constants import MAX_SEND_MESSAGE_LENGTH
from moddingway.util import log_info_and_add_field

logger = logging.getLogger(__name__)


async def send_message_to_channel(
    logging_embed: discord.Embed,
    channel: discord.TextChannel,
    message: str,
) -> str:
    if len(message) > MAX_SEND_MESSAGE_LENGTH:
        error_message = f"Message must be {MAX_SEND_MESSAGE_LENGTH} characters or less."
        log_info_and_add_field(logging_embed, logger, "Error", error_message)
        return error_message

    try:
        await channel.send(message)
    except discord.Forbidden:
        error_message = (
            f"Unable to post in {channel.mention}: the bot is missing permissions."
        )
        log_info_and_add_field(logging_embed, logger, "Error", error_message)
        return error_message
    except discord.HTTPException as error:
        error_message = f"Unable to post in {channel.mention}: {error}"
        log_info_and_add_field(logging_embed, logger, "Error", error_message)
        return error_message
    except Exception:
        error_message = (
            f"Unable to post in {channel.mention}: an unexpected error occurred."
        )
        log_info_and_add_field(logging_embed, logger, "Error", error_message)
        return error_message

    success_message = f"Successfully sent message to {channel.mention}"
    log_info_and_add_field(logging_embed, logger, "Result", success_message)
    return success_message
