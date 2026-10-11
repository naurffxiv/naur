from datetime import UTC, datetime, timedelta

import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import constants
from moddingway.constants import UserRole
from moddingway.database.models import User
from moddingway.services import exile_service
from moddingway.util import timestamp_to_epoch

DEFAULT_DATETIME_NOW = datetime(2019, 11, 19, 8, 0, 0, tzinfo=UTC)


def _mock_verified_exile_setup(
    mocker: MockerFixture, create_member, *, allows_dms=True
):
    mock_database_user = User(
        user_id=1,
        discord_user_id="12345",
        discord_guild_id="1",
        user_role=UserRole.USER,
        temporary_points=0,
        permanent_points=0,
        is_banned=False,
    )
    mocker.patch(
        "moddingway.database.exiles_database.get_user_active_exile", return_value=None
    )
    mocker.patch(
        "moddingway.database.users_database.get_user", return_value=mock_database_user
    )
    create_user_mock = mocker.patch(
        "moddingway.database.users_database.add_user", return_value=mock_database_user
    )
    exile_id = 4001
    mocker.patch("moddingway.database.exiles_database.add_exile", return_value=exile_id)
    mocked_add_note = mocker.patch(
        "moddingway.services.exile_service.note_service.add_note",
        new=mocker.AsyncMock(),
    )
    mocked_member = create_member(
        roles=[constants.Role.VERIFIED], allows_dms=allows_dms
    )
    mocked_logging_embed = mocker.Mock()
    return (
        mock_database_user,
        create_user_mock,
        exile_id,
        mocked_add_note,
        mocked_member,
        mocked_logging_embed,
    )


@pytest.mark.asyncio
async def test_exile_user__unverified(mocker: MockerFixture, create_member):
    # Arrange
    mocked_member = create_member()
    mock_database_user = User(
        user_id=1,
        discord_user_id="12345",
        discord_guild_id="1",
        user_role=UserRole.USER,
        temporary_points=0,
        permanent_points=0,
        is_banned=False,
    )
    mocker.patch(
        "moddingway.database.users_database.get_user", return_value=mock_database_user
    )
    mocker.patch(
        "moddingway.database.exiles_database.get_user_active_exile", return_value=None
    )
    mocker.patch("moddingway.util.user_has_role", return_value=False)
    mocked_add_note = mocker.patch(
        "moddingway.services.exile_service.note_service.add_note",
        new=mocker.AsyncMock(),
    )
    # Act
    res = await exile_service.exile_user(
        mocker.Mock(description=""),
        mocked_member,
        timedelta(days=1),
        "test_exile_user__unverified",
    )

    # Assert
    assert res is not None
    assert res == "User is not currently verified, no action will be taken"
    mocked_add_note.assert_not_called()


@pytest.mark.asyncio
async def test_exile_user__verified_existing_user_dm_failed(
    mocker: MockerFixture, create_member
):
    # Arrange
    (
        _mock_database_user,
        create_user_mock,
        exile_id,
        mocked_add_note,
        mocked_member,
        mocked_logging_embed,
    ) = _mock_verified_exile_setup(mocker, create_member, allows_dms=False)

    # Act
    res = await exile_service.exile_user(
        mocked_logging_embed,
        mocked_member,
        timedelta(days=1),
        "test_exile_user verified existing_user dm_failed",
    )

    # Assert
    assert res is None
    create_user_mock.assert_not_called()
    mocked_logging_embed.set_footer.assert_called_with(text=f"Exile ID: {exile_id}")
    mocked_add_note.assert_called_once()
    # TODO check exile create call to confirm data is correct

    assert any(
        call[1].get("name", "") == "DM Status"
        for call in mocked_logging_embed.add_field.call_args_list
    )


@pytest.mark.asyncio
async def test_exile_user__adds_note_on_exile(mocker: MockerFixture, create_member):
    # Arrange
    reason = "rule violation"
    duration = timedelta(days=1)
    end_timestamp = DEFAULT_DATETIME_NOW + duration
    timestamp = timestamp_to_epoch(end_timestamp)
    mocked_author = mocker.Mock(spec=discord.Member)
    mocked_author.id = 999

    (
        _mock_database_user,
        _create_user_mock,
        _exile_id,
        mocked_add_note,
        mocked_member,
        mocked_logging_embed,
    ) = _mock_verified_exile_setup(mocker, create_member)

    # Act
    await exile_service.exile_user(
        mocked_logging_embed,
        mocked_member,
        duration,
        reason,
        author=mocked_author,
    )

    # Assert
    mocked_add_note.assert_called_once_with(
        mocked_logging_embed,
        mocked_member,
        f"Exiled until <t:{timestamp}:F>. Reason: {reason}",
        mocked_author,
    )


@pytest.mark.asyncio
async def test_exile_user__roulette_does_not_add_note(
    mocker: MockerFixture, create_member
):
    # Arrange
    (
        _mock_database_user,
        _create_user_mock,
        _exile_id,
        mocked_add_note,
        mocked_member,
        mocked_logging_embed,
    ) = _mock_verified_exile_setup(mocker, create_member)

    # Act
    await exile_service.exile_user(
        mocked_logging_embed,
        mocked_member,
        timedelta(hours=6),
        "lost roulette",
        add_note=False,
    )

    # Assert
    mocked_add_note.assert_not_called()
