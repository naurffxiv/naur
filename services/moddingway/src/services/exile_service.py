import discord
import logging
from util import log_info_and_embed

logger = logging.getLogger(__name__)

async def exile_user(logging_embed: discord.Embed, user: discord.User, duration: str, reason: str):
    # TODO implement this functionality
    log_info_and_embed(logging_embed, logger, f"going to exile {user.mention} for {duration}")

    log_info_and_embed(logging_embed, logger, f"because {reason}")
    
    # look up user in DB

    # add exile entry into DB

    # change user role

    # message user?
    log_info_and_embed(logging_embed, logger, "Job's done")