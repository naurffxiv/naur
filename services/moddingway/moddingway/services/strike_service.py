import logging
from datetime import datetime, timedelta

import discord

from moddingway.constants import (
    MAX_STRIKE_REASON_LENGTH,
    MINOR_INFRACTION_POINTS,
    MODERATE_INFRACTION_POINTS,
    PERMANENT_BAN_STRIKE_THRESHOLD,
    SERIOUS_INFRACTION_POINTS,
    THRESHOLDS_PUNISHMENT,
    StrikeSeverity,
)
from moddingway.database import strikes_database, users_database
from moddingway.database.models import Strike, User
from moddingway.util import (
    log_info_and_add_field,
    log_info_and_embed,
    send_dm,
    timestamp_to_epoch,
)

from . import ban_service, exile_service

logger = logging.getLogger(__name__)


async def add_strike(
    logging_embed: discord.Embed,
    user: discord.Member,
    severity: StrikeSeverity,
    reason: str,
    author: discord.Member,
):
    # find user in DB
    db_user = users_database.get_user(user.id)
    if db_user is None:
        log_info_and_embed(
            logging_embed,
            logger,
            "User not found in database, creating new record",
        )
        db_user = users_database.add_user(user.id)

    # create strike
    strike_timestamp = datetime.now()
    strike = Strike(
        user_id=db_user.user_id,
        severity=severity,
        reason=reason,
        created_timestamp=strike_timestamp,
        created_by=str(author.id),
        last_edited_timestamp=strike_timestamp,
        last_edited_by=str(author.id),
    )
    strike.strike_id = strikes_database.add_strike(strike)
    logging_embed.set_footer(text=f"Strike ID: {strike.strike_id}")

    # increment user points, update
    db_user.last_infraction_timestamp = strike_timestamp
    previous_points = db_user.get_strike_points()
    _apply_strike_point_penalty(db_user, severity)
    users_database.update_user_strike_points(db_user)

    log_info_and_add_field(
        logging_embed,
        logger,
        "Strike Point Adjustment",
        f"<@{user.id}> was given a strike, bringing them to {db_user.get_strike_points()} points from {previous_points} points",
    )

    punishment = await _apply_punishment(logging_embed, user, db_user, previous_points)
    logging_embed.add_field(name="Punishment", value=punishment)

    # message user
    await send_dm(
        logging_embed,
        user,
        f"Your actions in NA Ultimate Raiding - FFXIV resulted in a strike against your account. This may result in punishment depending on the frequency or severity of your strikes.\n**Reason:** {reason}",
        context="Strike",
    )


async def get_user_strikes(
    user: discord.User,
) -> str:
    db_user = users_database.get_user(user.id)
    if db_user is None:
        return "User not found in database"

    strike_list = strikes_database.list_strikes(db_user.user_id)

    if len(strike_list) == 0:
        return "No strikes found for user"

    result = f"Strikes found for <@{user.id}>: [Temporary points: {db_user.temporary_points} | Permanent points: {db_user.permanent_points}]"
    for strike in strike_list:
        strike_id = strike[0]
        strike_severity = StrikeSeverity(strike[1])
        strike_reason = strike[2]
        strike_created_by = strike[3]
        strike_created_at = strike[4]
        strike_created_at_ts = timestamp_to_epoch(strike_created_at)
        result = (
            result
            + f"\n* ID: {strike_id} | SEVERITY: {strike_severity} | Moderator: <@{strike_created_by}> | DATE ISSUED: <t:{strike_created_at_ts}:F> | REASON: {strike_reason}"
        )

    result = result + f"\nTotal Points: {db_user.get_strike_points()}"

    return result


async def delete_strike(logging_embed: discord.Embed, strike_id: int) -> str:
    deleted_strike = strikes_database.delete_strike(strike_id)

    if deleted_strike is None:
        return "Strike not found, no change will be made"

    user_id = deleted_strike[1]
    severity = StrikeSeverity(int(deleted_strike[2]))
    temporary_points_to_remove = 0
    permanent_points_to_remove = 0

    if severity == StrikeSeverity.SERIOUS:
        permanent_points_to_remove = _get_severity_points(severity)
    else:
        temporary_points_to_remove = _get_severity_points(severity)

    users_database.decrement_user_strike_points(
        int(user_id), temporary_points_to_remove, permanent_points_to_remove
    )

    return f"Successfully deleted strike {strike_id}"


async def edit_strike(
    logging_embed: discord.Embed,
    editor: discord.User | discord.Member,
    client: discord.Client,
    strike_id: int,
    reason: str | None,
    severity: StrikeSeverity | None,
) -> str:
    strike = strikes_database.get_strike(strike_id)

    # doing this because of ruff's return count limit
    errors = []
    if strike is None:
        log_info_and_add_field(
            logging_embed,
            logger,
            "Result",
            "Strike not found",
        )
        errors.append("Strike not found in database")

    if reason is None and severity is None:
        errors.append("No updates provided. Please specify a reason and/or severity.")

    if reason is not None and len(reason) > MAX_STRIKE_REASON_LENGTH:
        errors.append(
            f"Reason is too long (max {MAX_STRIKE_REASON_LENGTH} characters). "
            "Please shorten the strike reason."
        )

    # first condition won't happen, but unresolved-attribute still complains
    if strike is None or len(errors) > 0:
        return "\n".join(errors)

    new_reason = reason if reason is not None else strike.reason
    old_severity = strike.severity
    new_severity = severity if severity is not None else old_severity

    if new_severity != old_severity:
        db_user = users_database.get_user_by_internal_id(strike.user_id)
        if db_user is None:
            return "User not found in database"

        temporary_points_to_remove = 0
        permanent_points_to_remove = 0

        if old_severity == StrikeSeverity.SERIOUS:
            permanent_points_to_remove = _get_severity_points(old_severity)
        else:
            temporary_points_to_remove = _get_severity_points(old_severity)

        users_database.decrement_user_strike_points(
            strike.user_id, temporary_points_to_remove, permanent_points_to_remove
        )

        # refetch after update
        db_user = users_database.get_user_by_internal_id(strike.user_id)
        if db_user is None:
            return "User not found in database"

        _apply_strike_point_penalty(db_user, new_severity)
        users_database.update_user_strike_points(db_user)

        log_info_and_add_field(
            logging_embed,
            logger,
            "Strike Point Adjustment",
            f"<@{strike.discord_user_id}> strike severity changed from {old_severity.name} to {new_severity.name}, bringing them to {db_user.get_strike_points()} points",
        )

    updated = strikes_database.update_strike(
        strike_id=strike_id,
        reason=new_reason,
        severity=new_severity,
        editor_id=str(editor.id),
        update_timestamp=datetime.now(),
    )

    if not updated:
        return "There was an error updating the strike"

    logging_embed.set_footer(text=f"Strike ID: {strike_id}")
    if reason is not None:
        logging_embed.add_field(name="Old reason", value=strike.reason)
    if severity is not None:
        logging_embed.add_field(name="Old severity", value=old_severity.name)

    log_info_and_add_field(
        logging_embed,
        logger,
        "Result",
        "Strike updated",
    )

    strike_user = await client.fetch_user(int(strike.discord_user_id))
    await send_dm(
        logging_embed,
        strike_user,
        "Your strike in NA Ultimate Raiding - FFXIV has been updated.\n"
        f"**Reason:** {new_reason}\n**Severity:** {new_severity.name}",
        context="Strike edit",
    )

    return "Strike successfully updated"


def _apply_strike_point_penalty(db_user: User, severity: StrikeSeverity):
    match severity:
        case StrikeSeverity.MINOR:
            db_user.temporary_points = (
                db_user.temporary_points + MINOR_INFRACTION_POINTS
            )
        case StrikeSeverity.MODERATE:
            db_user.temporary_points = (
                db_user.temporary_points + MODERATE_INFRACTION_POINTS
            )
        case StrikeSeverity.SERIOUS:
            db_user.permanent_points = (
                db_user.permanent_points + SERIOUS_INFRACTION_POINTS
            )


def _get_severity_points(severity: StrikeSeverity) -> int:
    match severity:
        case StrikeSeverity.MINOR:
            return MINOR_INFRACTION_POINTS
        case StrikeSeverity.MODERATE:
            return MODERATE_INFRACTION_POINTS
        case StrikeSeverity.SERIOUS:
            return SERIOUS_INFRACTION_POINTS


def _calculate_punishment(previous_points: int, total_points: int) -> int:
    punishment_days = 0
    jumped_multiple_thresholds = False

    for points_threshold, days in THRESHOLDS_PUNISHMENT:
        if total_points >= points_threshold and previous_points < points_threshold:
            punishment_days += days
            jumped_multiple_thresholds = True

    if not jumped_multiple_thresholds:
        for points_threshold, days in THRESHOLDS_PUNISHMENT:
            if total_points >= points_threshold and previous_points >= points_threshold:
                punishment_days = days
                break

    return punishment_days


async def _apply_punishment(
    logging_embed: discord.Embed,
    user: discord.Member,
    db_user: User,
    previous_points: int,
) -> str:
    total_points = db_user.get_strike_points()

    # TODO: MOD-93 known error, if an exiled user is given a strike, the follow up exile is not created
    exile_reason = (
        "Your actions were severe or frequent enough for you to receive this exile"
    )

    if total_points >= PERMANENT_BAN_STRIKE_THRESHOLD:
        punishment = "Permanent ban"
        await ban_service.ban_member(
            logging_embed,
            user,
            "Your strike were severe or frequent to be removed from NA Ultimate Raiding - FFXIV",
            False,
        )
        return punishment

    punishment_days = _calculate_punishment(previous_points, total_points)

    if punishment_days == 0:
        punishment = "Nothing"
    else:
        punishment = f"{punishment_days} day exile"
        await exile_service.exile_user(
            logging_embed, user, timedelta(days=punishment_days), exile_reason
        )
    return punishment
