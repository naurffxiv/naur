import discord
import logging
from util import log_info_and_embed, add_and_remove_role, send_dm
from enums import Role
from database import users_database
from typing import Optional

logger = logging.getLogger(__name__)


async def exile_user(
    logging_embed: discord.Embed,
    user: discord.Member,
    duration: Optional[str],
    reason: str,
):
    # TODO implement this functionality

    if duration:
        # look up user in DB
        db_user = users_database.get_user(user.id)
        if db_user is None:
            log_info_and_embed(
                logging_embed,
                logger,
                f"User not found in database, creating new record",
            )
            users_database.add_user(user.id)
            log_info_and_embed(
                logging_embed, logger, f"User record created in database"
            )

        # add exile entry into DB

    # change user role
    await add_and_remove_role(
        user, role_to_add=Role.EXILED, role_to_remove=Role.VERIFIED
    )

    # message user
    await send_dm(
        user, f"You are being exiled from NAUR FFXIV for the following reason: {reason}"
    )


async def unexile_user(logging_embed: discord.Embed, user: discord.User):

    await add_and_remove_role(user, Role.VERIFIED, Role.EXILED)
