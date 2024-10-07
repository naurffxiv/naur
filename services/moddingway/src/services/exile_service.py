import discord
import logging
from util import log_info_and_embed, add_and_remove_role, send_dm
from enums import Role
from database import users_database, exiles_database, DatabaseConnection
from typing import Optional
import datetime
from database.models import Exile, User

logger = logging.getLogger(__name__)


async def exile_user(
    logging_embed: discord.Embed,
    user: discord.Member,
    duration: Optional[datetime.timedelta],
    reason: str,
):
    if duration:
        # look up user in DB
        db_user = users_database.get_user(user.id)
        if db_user is None:
            log_info_and_embed(
                logging_embed,
                logger,
                f"User not found in database, creating new record",
            )
            db_user_id = users_database.add_user(user.id)
            logger.info(f"Created user record in DB with id {db_user_id}")

            db_user = User(db_user_id, user.id, None, None)

        # add exile entry into DB
        start_timestamp = datetime.datetime.now(datetime.timezone.utc)
        end_timestamp = datetime.datetime.now(datetime.timezone.utc) + duration
        exile = Exile(None, db_user.user_id, reason, 1, start_timestamp, end_timestamp)
        exile_id = exiles_database.add_exile(exile)
        log_info_and_embed(logging_embed, logger, f"Created exile with ID {exile_id}")

    # change user role
    await add_and_remove_role(
        user, role_to_add=Role.EXILED, role_to_remove=Role.VERIFIED
    )

    # message user
    await send_dm(
        user, f"You are being exiled from NAUR FFXIV for the following reason: {reason}"
    )


async def unexile_user(logging_embed: discord.Embed, user: discord.User):

    db_user = users_database.get_user(user.id)
    if db_user is None:
        log_info_and_embed(
            logging_embed,
            logger,
            f"User not found in database, creating new record",
        )
        users_database.add_user(user.id)
        log_info_and_embed(logging_embed, logger, f"User record created in database")

    await add_and_remove_role(user, Role.VERIFIED, Role.EXILED)
