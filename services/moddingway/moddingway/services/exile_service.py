import logging
from dataclasses import dataclass

import discord

from moddingway.constants import ExileStatus, Role
from moddingway.database import exiles_database, users_database
from moddingway.settings import get_settings
from moddingway.util import (
    log_info_and_add_field,
    log_info_and_embed,
    parse_duration,
    timestamp_to_epoch,
)

from .exiles import ExileOutcome, ExileResult

settings = get_settings()
logger = logging.getLogger(__name__)

NOT_VERIFIED_MESSAGE = "User is not currently verified, no action will be taken"
NOT_EXILED_MESSAGE = "User is not currently exiled, no action will be taken"
MISSING_USER_RECORD_MESSAGE = (
    "User has been unexiled, but no user record was found in the database"
)


@dataclass(frozen=True)
class _OutcomeText:
    # "Error" field, logged before any other problem
    error: str | None = None
    # Fields logged last, in order
    fields: tuple[tuple[str, str], ...] = ()
    # Reply to the moderator; defaults to ``error``, then the success message
    reply: str | None = None


def _outcome_text(result: ExileResult) -> _OutcomeText:
    """Every message the bot shows for each Exiles outcome."""
    mention = f"<@{result.discord_id}>"
    end_epoch = timestamp_to_epoch(result.end_timestamp)
    expiration = ("Expiration", f"<t:{end_epoch}:R>")
    table = {
        ExileOutcome.EXILED: _OutcomeText(
            fields=(expiration, ("Result", f"{mention} was successfully exiled"))
        ),
        ExileOutcome.EXTENDED: _OutcomeText(
            fields=(("Result", f"Exile extended until <t:{end_epoch}>"),),
            reply="User exile extended",
        ),
        ExileOutcome.NOT_VERIFIED: _OutcomeText(error=NOT_VERIFIED_MESSAGE),
        ExileOutcome.UNEXILED: _OutcomeText(
            fields=(("Result", f"{mention} was successfully unexiled"),)
        ),
        ExileOutcome.NOT_EXILED: _OutcomeText(error=NOT_EXILED_MESSAGE),
        ExileOutcome.MEMBER_GONE: _OutcomeText(
            fields=(
                (
                    "Result",
                    f"{mention} was not found in the server, exile marked as ended without role changes",
                ),
            )
        ),
        ExileOutcome.ALREADY_RELEASED: _OutcomeText(
            fields=(
                (
                    "Result",
                    f"{mention} no longer had the {Role.EXILED.value} role, exile marked as ended without role changes",
                ),
            )
        ),
        ExileOutcome.REAPPLIED: _OutcomeText(
            fields=(
                (
                    "Result",
                    f"{mention} rejoined during an active exile and was given the {Role.EXILED.value} role instead of {Role.NON_VERIFIED.value}",
                ),
                expiration,
            )
        ),
        ExileOutcome.NO_ACTIVE_EXILE: _OutcomeText(),
        ExileOutcome.FAILED: _OutcomeText(
            error=result.error or "Unknown error",
            reply=result.error or "An error occurred while processing the command.",
        ),
    }
    return table[result.outcome]


def log_exile_result(logging_embed: discord.Embed, result: ExileResult):
    """Describe an Exiles result on a logging embed, as the bot has always done."""
    if result.created_user_record:
        log_info_and_embed(
            logging_embed,
            logger,
            "User not found in database, creating new record",
        )

    # Extensions have never shown the exile ID
    if result.exile_id is not None and result.outcome != ExileOutcome.EXTENDED:
        logging_embed.set_footer(text=f"Exile ID: {result.exile_id}")

    text = _outcome_text(result)
    if text.error is not None:
        log_info_and_add_field(logging_embed, logger, "Error", text.error)
    for problem in result.problems:
        log_info_and_add_field(logging_embed, logger, "Error", problem)
    if result.missing_user_record:
        log_info_and_add_field(
            logging_embed, logger, "Error", MISSING_USER_RECORD_MESSAGE
        )
    if result.dm_error:
        log_info_and_add_field(logging_embed, logger, "DM Status", result.dm_error)
    for name, value in text.fields:
        log_info_and_add_field(logging_embed, logger, name, value)


def exile_command_response(result: ExileResult, success_message: str) -> str:
    """The reply shown to the moderator who ran /exile or /unexile."""
    text = _outcome_text(result)
    if text.reply is not None:
        return text.reply
    if text.error is not None:
        return text.error
    if result.missing_user_record:
        return MISSING_USER_RECORD_MESSAGE
    return success_message


async def delete_exile_by_id(logging_embed: discord.Embed, exile_id: int):
    status = exiles_database.get_exile_status(exile_id)
    if status == ExileStatus.TIMED_EXILED:
        log_info_and_add_field(
            logging_embed,
            logger,
            "Result",
            "Selected exile is active, cannot be deleted",
        )
        return "Selected exile is active, cannot be deleted"
    elif status == ExileStatus.UNEXILED:
        result = exiles_database.delete_exile(exile_id)
        if result:
            log_info_and_add_field(
                logging_embed,
                logger,
                "Result",
                f"Exile {exile_id} successfully deleted",
            )
            return f"Exile {exile_id} successfully deleted"
        else:
            log_info_and_add_field(
                logging_embed,
                logger,
                "Result",
                f"Error with exile {exile_id} deletion",
            )
            return f"Error with exile {exile_id} deletion"
    elif not status:
        log_info_and_add_field(
            logging_embed, logger, "Result", f"Exile {exile_id} not found"
        )
        return f"Exile {exile_id} not found"


TIME_FORMAT = "%Y-%m-%d %H:%M:%S"


async def get_user_exiles(user: discord.User) -> str:
    db_user = users_database.get_user(user.id)
    if db_user is None:
        return "User not found in database"

    exile_list = exiles_database.get_user_exiles(db_user.user_id)

    if len(exile_list) == 0:
        return "No exiles found for user"
    else:
        result = f"Exiles found for <@{user.id}>:"
        for exile in exile_list:
            exile_id = exile.exile_id
            exile_reason = exile.reason
            exile_start_epoch = timestamp_to_epoch(exile.start_timestamp)
            exile_end_epoch = (
                timestamp_to_epoch(exile.end_timestamp) if exile.exile_status else None
            )
            exile_start_date = f"<t:{exile_start_epoch}:F>"
            exile_end_date = (
                f"<t:{exile_end_epoch}:F>" if exile_end_epoch else "Indefinite"
            )
            exile_type = ExileStatus(exile.exile_status).name

            result = (
                result
                + f"\n* ID: {exile_id} | START DATE: {exile_start_date} | END DATE: {exile_end_date} | TYPE: {exile_type} | REASON: {exile_reason}"
            )

    return result


async def get_active_exiles() -> str:
    exile_list = exiles_database.get_all_active_exiles()
    logger.info("Database query completed.")
    if len(exile_list) == 0:
        return "No active exiles found"

    result = "Active exiles found for"
    for exile in exile_list:
        logger.info(type(exile))
        exile_id = exile.exile_id
        discord_id = f"<@{exile.discord_id}>"
        exile_reason = exile.reason
        exile_start_epoch = timestamp_to_epoch(exile.start_timestamp)
        exile_end_epoch = (
            timestamp_to_epoch(exile.end_timestamp) if exile.exile_status else None
        )
        exile_start_date = f"<t:{exile_start_epoch}:F>"
        exile_end_date = f"<t:{exile_end_epoch}:F>" if exile_end_epoch else "Indefinite"
        exile_type = ExileStatus(exile.exile_status).name

        result = (
            result
            + f"\n* ID: {exile_id} | USER: {discord_id} | START DATE: {exile_start_date} | END DATE: {exile_end_date} | TYPE: {exile_type} | REASON: {exile_reason}"
        )

    return result


def format_time_string(duration_string: str):
    parsed = parse_duration(duration_string)
    if parsed is None:
        return None

    amount, unit = parsed
    return f"{amount} {unit}{'' if amount == 1 else 's'}"
