from datetime import UTC, datetime, timedelta
from types import SimpleNamespace

import discord

from moddingway.constants import ExileStatus, Role, StrikeSeverity
from moddingway.services import exile_service, exiles, strike_service
from moddingway.services.exile_service import (
    MISSING_USER_RECORD_MESSAGE,
    exile_command_response,
    log_exile_result,
)
from moddingway.services.exiles import ExileOutcome
from moddingway.workers.autounexile import autounexile_users

from .conftest import EXILED_ROLE_ID, MOD_ROLE_ID, STICKY_ROLE_IDS, VERIFIED_ROLE_ID

STICKY_ROLE = STICKY_ROLE_IDS[0]


def exile_rows(query) -> list[tuple]:
    return query(
        """
        SELECT e.exileID, u.discordUserID, e.exileStatus, e.endTimestamp
        FROM exiles e JOIN users u ON e.userID = u.userID
        ORDER BY e.exileID
        """
    )


def sticky_rows(query) -> list[str]:
    return [row[0] for row in query("SELECT roleID FROM roles ORDER BY entryID")]


def as_utc(timestamp: datetime) -> datetime:
    return timestamp.replace(tzinfo=UTC)


def after_end_of_exile() -> datetime:
    return datetime.now(UTC) + timedelta(days=30)


# Exile


async def test_exile__new_member_is_saved_and_exiled(make_member, query):
    member = make_member(role_ids=[STICKY_ROLE])

    result = await exiles.exile(member, timedelta(days=1), "spamming")

    assert result.outcome == ExileOutcome.EXILED
    assert result.created_user_record
    assert result.error is None and result.problems == [] and result.dm_error is None

    [(exile_id, discord_id, status, end)] = exile_rows(query)
    assert exile_id == result.exile_id
    assert discord_id == str(member.id)
    assert status == ExileStatus.TIMED_EXILED
    assert as_utc(end) == result.end_timestamp
    assert sticky_rows(query) == [str(STICKY_ROLE)]

    assert member.role_ids() == {EXILED_ROLE_ID}
    [dm] = member.dm_channel.sent
    assert "You are being exiled" in dm and "spamming" in dm

    embed = discord.Embed()
    log_exile_result(embed, result)
    assert [field.name for field in embed.fields] == ["Expiration", "Result"]
    assert embed.fields[1].value == f"<@{member.id}> was successfully exiled"
    assert embed.footer.text == f"Exile ID: {exile_id}"
    assert embed.description is not None
    assert "creating new record" in embed.description


async def test_exile__already_exiled_member_is_extended(make_member, query):
    member = make_member()
    first = await exiles.exile(member, timedelta(days=1), "first")
    assert first.end_timestamp is not None

    second = await exiles.exile(member, timedelta(days=2), "second")

    assert second.outcome == ExileOutcome.EXTENDED
    assert second.exile_id == first.exile_id
    [(_, _, status, end)] = exile_rows(query)
    assert status == ExileStatus.TIMED_EXILED
    assert as_utc(end) == first.end_timestamp + timedelta(days=2)
    assert member.role_ids() == {EXILED_ROLE_ID}
    assert len(member.dm_channel.sent) == 1


async def test_exile__unverified_member_is_refused(make_member, query):
    member = make_member(roles=[Role.NON_VERIFIED])

    result = await exiles.exile(member, timedelta(days=1), "reason")

    assert result.outcome == ExileOutcome.NOT_VERIFIED
    assert exile_rows(query) == []
    assert query("SELECT * FROM users") == []


async def test_exile__role_swap_failure_rolls_back_everything(make_member, query):
    member = make_member()
    member.failing_role_ids.add(EXILED_ROLE_ID)

    result = await exiles.exile(member, timedelta(days=1), "reason")

    assert result.outcome == ExileOutcome.FAILED
    assert result.error is not None and "no changes were saved" in result.error
    # Neither the new user record nor the exile was committed
    assert query("SELECT * FROM users") == []
    assert exile_rows(query) == []
    # The member keeps the role they had and is not messaged
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    assert member.dm_channel.sent == []


async def test_exile__sticky_role_failure_is_reported_but_not_rolled_back(
    make_member, query
):
    member = make_member(role_ids=[STICKY_ROLE])
    member.failing_role_ids.add(STICKY_ROLE)

    result = await exiles.exile(member, timedelta(days=1), "reason")

    assert result.outcome == ExileOutcome.EXILED
    [problem] = result.problems
    assert problem.startswith("Sticky roles could not be completely removed")
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.TIMED_EXILED
    assert member.role_ids() == {EXILED_ROLE_ID, STICKY_ROLE}
    assert sticky_rows(query) == []


async def test_exile_and_unexile__other_roles_are_kept(make_member):
    member = make_member(roles=[Role.VERIFIED, Role.MOD])

    await exiles.exile(member, timedelta(days=1), "reason")
    assert member.role_ids() == {EXILED_ROLE_ID, MOD_ROLE_ID}

    await exiles.unexile(member)
    assert member.role_ids() == {VERIFIED_ROLE_ID, MOD_ROLE_ID}


async def test_exile__dm_failure_is_reported(make_member, query):
    member = make_member()
    member.dms_open = False

    result = await exiles.exile(member, timedelta(days=1), "reason")

    assert result.outcome == ExileOutcome.EXILED
    assert result.dm_error is not None
    assert len(exile_rows(query)) == 1


# Unexile


async def test_unexile__ends_exile_and_restores_sticky_roles(make_member, query):
    member = make_member(role_ids=[STICKY_ROLE])
    exiled = await exiles.exile(member, timedelta(days=1), "reason")

    result = await exiles.unexile(member)

    assert result.outcome == ExileOutcome.UNEXILED
    assert result.exile_id == exiled.exile_id
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNEXILED
    assert member.role_ids() == {VERIFIED_ROLE_ID, STICKY_ROLE}
    assert sticky_rows(query) == []
    assert "un-exiled" in member.dm_channel.sent[-1]


async def test_unexile__member_without_user_record_is_unexiled(make_member, query):
    member = make_member(roles=[Role.EXILED])

    result = await exiles.unexile(member)

    assert result.outcome == ExileOutcome.UNEXILED
    assert result.missing_user_record
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    assert query("SELECT * FROM users") == []
    assert exile_command_response(result, "ok") == MISSING_USER_RECORD_MESSAGE
    embed = discord.Embed()
    log_exile_result(embed, result)
    assert [(f.name, f.value) for f in embed.fields] == [
        ("Error", MISSING_USER_RECORD_MESSAGE),
        ("Result", f"<@{member.id}> was successfully unexiled"),
    ]


async def test_unexile__role_swap_failure_rolls_back(make_member, query):
    member = make_member()
    await exiles.exile(member, timedelta(days=1), "reason")
    member.failing_role_ids.add(VERIFIED_ROLE_ID)

    result = await exiles.unexile(member)

    assert result.outcome == ExileOutcome.FAILED
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.TIMED_EXILED
    assert member.role_ids() == {EXILED_ROLE_ID}


async def test_unexile__sticky_role_failure_is_reported_but_not_rolled_back(
    make_member, query
):
    member = make_member(role_ids=[STICKY_ROLE])
    await exiles.exile(member, timedelta(days=1), "reason")
    member.failing_role_ids.add(STICKY_ROLE)

    result = await exiles.unexile(member)

    assert result.outcome == ExileOutcome.UNEXILED
    [problem] = result.problems
    assert problem.startswith("Sticky roles could not be completely restored")
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNEXILED
    assert member.role_ids() == {VERIFIED_ROLE_ID}
    # Kept so the roles can still be restored by hand
    assert sticky_rows(query) == [str(STICKY_ROLE)]


# Release of due exiles (autounexile worker)


async def test_release_due__releases_only_finished_exiles(guild, make_member, query):
    due = make_member(member_id=1)
    running = make_member(member_id=2)
    await exiles.exile(due, timedelta(seconds=1), "short")
    await exiles.exile(running, timedelta(days=60), "long")

    results = await exiles.release_due(guild, datetime.now(UTC) + timedelta(days=1))

    assert [(r.discord_id, r.outcome) for r in results] == [(1, ExileOutcome.UNEXILED)]
    assert [row[2] for row in exile_rows(query)] == [
        ExileStatus.UNEXILED,
        ExileStatus.TIMED_EXILED,
    ]
    assert due.role_ids() == {VERIFIED_ROLE_ID}
    assert running.role_ids() == {EXILED_ROLE_ID}


async def test_release_due__member_gone_is_marked_unknown(guild, make_member, query):
    member = make_member()
    await exiles.exile(member, timedelta(days=1), "reason")
    del guild.members[member.id]

    [result] = await exiles.release_due(guild, after_end_of_exile())

    assert result.outcome == ExileOutcome.FAILED
    assert result.member is None
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNKNOWN


async def test_release_due__failed_unexile_is_marked_unknown(guild, make_member, query):
    member = make_member()
    await exiles.exile(member, timedelta(days=1), "reason")
    member.failing_role_ids.add(VERIFIED_ROLE_ID)

    [result] = await exiles.release_due(guild, after_end_of_exile())

    assert result.outcome == ExileOutcome.FAILED
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNKNOWN
    assert member.role_ids() == {EXILED_ROLE_ID}
    # Not picked up again
    assert await exiles.release_due(guild, after_end_of_exile()) == []


async def test_release_due__member_already_without_exiled_role(
    guild, make_member, query
):
    member = make_member()
    await exiles.exile(member, timedelta(days=1), "reason")
    member.roles = [guild.role(Role.VERIFIED)]  # a mod removed Exiled by hand

    [result] = await exiles.release_due(guild, after_end_of_exile())

    assert result.outcome == ExileOutcome.FAILED
    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNKNOWN


async def test_autounexile_worker__releases_due_exiles(guild, make_member, query):
    member = make_member()
    await exiles.exile(member, timedelta(seconds=-1), "already over")
    bot = SimpleNamespace(get_guild=lambda _guild_id: guild, get_channel=lambda _: None)

    assert await autounexile_users(bot) == "Auto Unexile task completed."

    [(_, _, status, _)] = exile_rows(query)
    assert status == ExileStatus.UNEXILED
    assert member.role_ids() == {VERIFIED_ROLE_ID}


# Strike escalation


async def test_strike_escalation__strike_saved_then_member_exiled(make_member, query):
    member = make_member()
    moderator = make_member(member_id=999, roles=[Role.MOD])
    embed = discord.Embed()

    await strike_service.add_strike(
        embed, member, StrikeSeverity.MODERATE, "rude", moderator
    )

    [(strike_reason,)] = query("SELECT reason FROM strikes")
    assert strike_reason == "rude"
    [(_, _, status, end)] = exile_rows(query)
    assert status == ExileStatus.TIMED_EXILED
    assert as_utc(end) - datetime.now(UTC) > timedelta(hours=23)
    assert member.role_ids() == {EXILED_ROLE_ID}
    punishment = next(f for f in embed.fields if f.name == "Punishment")
    assert punishment.value == "1 day exile"


async def test_strike_escalation__strike_kept_when_exile_fails(make_member, query):
    member = make_member()
    member.failing_role_ids.add(EXILED_ROLE_ID)
    moderator = make_member(member_id=999, roles=[Role.MOD])
    embed = discord.Embed()

    await strike_service.add_strike(
        embed, member, StrikeSeverity.MODERATE, "rude", moderator
    )

    assert len(query("SELECT * FROM strikes")) == 1
    assert exile_rows(query) == []
    assert any(
        f.name == "Error" and "no changes were saved" in str(f.value)
        for f in embed.fields
    )


# Exile history commands


async def test_exile_history__lists_and_deletes_ended_exiles(make_member, query):
    member = make_member()
    exiled = await exiles.exile(member, timedelta(days=1), "spamming")

    active = await exile_service.get_active_exiles()
    assert f"ID: {exiled.exile_id} | USER: <@{member.id}>" in active
    history = await exile_service.get_user_exiles(member)
    assert f"ID: {exiled.exile_id}" in history and "REASON: spamming" in history
    embed = discord.Embed()
    assert (
        await exile_service.delete_exile_by_id(embed, exiled.exile_id)
        == "Selected exile is active, cannot be deleted"
    )

    await exiles.unexile(member)

    assert await exile_service.get_active_exiles() == "No active exiles found"
    assert (
        await exile_service.delete_exile_by_id(embed, exiled.exile_id)
        == f"Exile {exiled.exile_id} successfully deleted"
    )
    assert exile_rows(query) == []
    assert await exile_service.get_user_exiles(member) == "No exiles found for user"
    assert (
        await exile_service.delete_exile_by_id(embed, exiled.exile_id)
        == f"Exile {exiled.exile_id} not found"
    )


async def test_exile_history__unknown_user(make_member):
    assert (
        await exile_service.get_user_exiles(make_member())
        == "User not found in database"
    )
