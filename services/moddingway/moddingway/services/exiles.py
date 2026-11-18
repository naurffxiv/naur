"""
Exiles: the single owner of the Exile lifecycle.

Every way an Exile starts or ends goes through this module:

* ``exile`` - /exile, /roulette and strike escalation
* ``unexile`` - the manual /unexile command
* ``release_due`` - the autounexile worker

Each call returns an ``ExileResult`` describing what happened. Callers build
their own logging embed from it (see ``exile_service.log_exile_result``); this
module never touches an embed and does not raise for expected failures.

The database writes for an Exile/Unexile and the Exiled/Verified role swap share
one Postgres transaction that is only committed once the role swap succeeded.
The swap is a single Discord call, so it is never half applied. Sticky roles and
DMs are best-effort and happen after the commit.
"""

import logging
from dataclasses import dataclass, field
from datetime import UTC, datetime, timedelta
from enum import StrEnum

import discord

from moddingway.constants import ExileStatus, Role
from moddingway.database import (
    DatabaseConnection,
    exiles_database,
    roles_database,
    users_database,
)
from moddingway.database.models import Exile, PendingExile
from moddingway.settings import get_settings
from moddingway.util import (
    add_and_remove_role,
    timestamp_to_epoch,
    try_send_dm,
    user_has_role,
)

settings = get_settings()
logger = logging.getLogger(__name__)


class ExileOutcome(StrEnum):
    # Exile
    EXILED = "exiled"
    EXTENDED = "extended"
    NOT_VERIFIED = "not_verified"
    # Unexile
    UNEXILED = "unexiled"
    NOT_EXILED = "not_exiled"
    # Any operation: nothing was saved
    FAILED = "failed"


@dataclass
class ExileResult:
    outcome: ExileOutcome
    discord_id: int
    member: discord.Member | None = None
    exile_id: int | None = None
    end_timestamp: datetime | None = None
    created_user_record: bool = False
    # Set when a member was unexiled but had no user record in the database
    missing_user_record: bool = False
    # Set when the operation did not happen; no DB changes were committed
    error: str | None = None
    # Non-fatal problems (e.g. sticky roles) worth surfacing to mods
    problems: list[str] = field(default_factory=list)
    # Set when the member could not be messaged
    dm_error: str | None = None


def _result(
    outcome: ExileOutcome,
    discord_id: int,
    member: discord.Member | None,
    exile: PendingExile | None,
    **kwargs,
) -> ExileResult:
    """An ExileResult about ``exile`` (when there is one)."""
    return ExileResult(
        outcome,
        discord_id,
        member,
        exile_id=exile.exile_id if exile is not None else None,
        end_timestamp=exile.end_timestamp if exile is not None else None,
        **kwargs,
    )


async def exile(
    member: discord.Member, duration: timedelta, reason: str
) -> ExileResult:
    """Exile ``member`` for ``duration``, or extend their current Exile."""
    user_id, current_exile = exiles_database.get_user_and_active_exile(member.id)
    if current_exile is not None and not user_has_role(member, Role.VERIFIED):
        current_exile.end_timestamp += duration
        exiles_database.update_exile_end(
            current_exile.exile_id, current_exile.end_timestamp
        )
        logger.info(
            f"Exile {current_exile.exile_id} extended until {current_exile.end_timestamp}"
        )
        return _result(ExileOutcome.EXTENDED, member.id, member, current_exile)

    if not user_has_role(member, Role.VERIFIED):
        return ExileResult(ExileOutcome.NOT_VERIFIED, member.id, member)

    start_timestamp = datetime.now(UTC)
    end_timestamp = start_timestamp + duration
    created_user_record = user_id is None

    try:
        with DatabaseConnection().transaction():
            if user_id is None:
                user_id = users_database.add_user(member.id).user_id
            exile_id = exiles_database.add_exile(
                Exile(
                    exile_id=None,
                    user_id=user_id,
                    discord_id=member.id,
                    reason=reason,
                    exile_status=ExileStatus.TIMED_EXILED.value,
                    start_timestamp=start_timestamp,
                    end_timestamp=end_timestamp,
                )
            )
            # Last step, so a failure rolls the transaction back
            await add_and_remove_role(
                member, role_to_add=Role.EXILED, role_to_remove=Role.VERIFIED
            )
    except Exception as e:
        logger.exception(f"Failed to exile {member.id}, no changes were saved")
        return ExileResult(
            ExileOutcome.FAILED,
            member.id,
            member,
            error=f"Exile failed and no changes were saved: {e}",
        )

    logger.info(f"Created exile with ID {exile_id}")
    result = ExileResult(
        ExileOutcome.EXILED,
        member.id,
        member,
        exile_id=exile_id,
        end_timestamp=end_timestamp,
        created_user_record=created_user_record,
    )

    await _strip_sticky_roles(member, user_id, result)

    result.dm_error = await try_send_dm(
        member,
        f"You are being exiled from NA Ultimate Raiding - FFXIV.\n**Reason:** {reason}\nExile expiration: <t:{timestamp_to_epoch(end_timestamp)}:R>",
        context="Exile",
    )
    return result


async def unexile(member: discord.Member) -> ExileResult:
    """End ``member``'s Exile now (the manual /unexile command)."""
    if not user_has_role(member, Role.EXILED):
        return ExileResult(ExileOutcome.NOT_EXILED, member.id, member)

    user_id, active_exile = exiles_database.get_user_and_active_exile(member.id)
    if user_id is not None and active_exile is None:
        # This should only happen when someone was manually exiled then
        # unexiled through the bot
        logger.info(
            f"No active exile associated with user ID {user_id} was found. Skipping DB actions..."
        )

    result = await _unexile_member(member, user_id, active_exile)
    result.missing_user_record = (
        user_id is None and result.outcome == ExileOutcome.UNEXILED
    )
    return result


async def release_due(guild: discord.Guild, now: datetime) -> list[ExileResult]:
    """
    End every Exile whose end time is before ``now``.

    An Exile that cannot be ended (the member left the server, no longer has the
    Exiled role, or the Unexile failed) is marked UNKNOWN, so it is not picked up
    again.
    """
    results: list[ExileResult] = []
    for pending_exile in exiles_database.get_pending_unexiles(now):
        logger.info(f"Releasing exile {pending_exile.exile_id}.")
        try:
            result = await _release(guild, pending_exile)
        except Exception as e:
            logger.exception(f"Failed to release exile {pending_exile.exile_id}")
            result = _result(
                ExileOutcome.FAILED,
                pending_exile.discord_id,
                None,
                pending_exile,
                error=f"Unexile failed and no changes were saved: {e}",
            )
        if result.outcome == ExileOutcome.FAILED:
            exiles_database.update_exile_status(
                pending_exile.exile_id, ExileStatus.UNKNOWN
            )
        results.append(result)
    return results


async def _release(guild: discord.Guild, pending_exile: PendingExile) -> ExileResult:
    member = guild.get_member(pending_exile.discord_id)
    if member is None:
        error = f"<@{pending_exile.discord_id}> was not found in the server."
    elif not user_has_role(member, Role.EXILED):
        error = "User is not currently exiled, no action will be taken"
    else:
        return await _unexile_member(member, pending_exile.user_id, pending_exile)

    return _result(
        ExileOutcome.FAILED,
        pending_exile.discord_id,
        member,
        pending_exile,
        error=error,
    )


async def _unexile_member(
    member: discord.Member,
    user_id: int | None,
    active_exile: PendingExile | None,
) -> ExileResult:
    try:
        with DatabaseConnection().transaction():
            if active_exile is not None:
                exiles_database.update_exile_status(
                    active_exile.exile_id, ExileStatus.UNEXILED
                )
            # Last step, so a failure rolls the transaction back
            await add_and_remove_role(
                member, role_to_add=Role.VERIFIED, role_to_remove=Role.EXILED
            )
    except Exception as e:
        logger.exception(f"Failed to unexile {member.id}, no changes were saved")
        return _result(
            ExileOutcome.FAILED,
            member.id,
            member,
            active_exile,
            error=f"Unexile failed and no changes were saved: {e}",
        )

    result = _result(ExileOutcome.UNEXILED, member.id, member, active_exile)

    if user_id is not None:
        await _restore_sticky_roles(member, user_id, result)

    result.dm_error = await try_send_dm(
        member,
        "You have been un-exiled from NA Ultimate Raiding - FFXIV.",
        context="Un-Exile",
    )
    return result


async def _strip_sticky_roles(
    member: discord.Member, user_id: int, result: ExileResult
):
    try:
        roles_to_save: list[discord.Role] = []
        for role in member.roles:
            if role.id in settings.sticky_roles:
                discord_role = member.guild.get_role(role.id)
                if discord_role is not None:
                    roles_to_save.append(discord_role)

        if len(roles_to_save) > 0:
            await member.remove_roles(*roles_to_save)
            roles_database.add_sticky_roles(
                user_id, [role.id for role in roles_to_save]
            )
    except Exception as e:
        logger.exception("Failed to strip sticky roles")
        result.problems.append(f"Sticky roles could not be completely removed, {e}")


async def _restore_sticky_roles(
    member: discord.Member, user_id: int, result: ExileResult
):
    try:
        roles_to_restore: list[discord.Role] = []
        for role in roles_database.get_sticky_roles(user_id) or []:
            discord_role = member.guild.get_role(int(role))
            if discord_role is not None:
                roles_to_restore.append(discord_role)
            else:
                logger.error("Role %s was not found in the server, skipping...", role)

        if len(roles_to_restore) > 0:
            await member.add_roles(*roles_to_restore)
            roles_database.remove_sticky_roles(user_id)
    except Exception as e:
        logger.exception("Failed to restore sticky roles")
        result.problems.append(f"Sticky roles could not be completely restored, {e}")
