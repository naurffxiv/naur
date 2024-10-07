import discord
import logging
from util import log_info_and_embed, send_dm

logger = logging.getLogger(__name__)


async def ban_user(logging_embed: discord.Embed, user: discord.Member, reason: str):
    log_info_and_embed(logging_embed, logger, f"going to ban {user.mention}")

    log_info_and_embed(logging_embed, logger, f"because {reason}")

    await send_dm(
        user,
        f"You are being banned from NA Ultimate Raiding - FF XIV for the following reason: \n> {reason} \
        \nYou are able to appeal this ban through the link provided: https://dyno.gg/form/33b5a650",
    )

    await user.ban(reason=reason)
