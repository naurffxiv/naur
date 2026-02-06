from datetime import timedelta
from typing import List

import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import constants, util

single_line_case = ("short message", 100, ["short message"])
medium_len_case = (
    "medium length message that gets broken in half",
    35,
    ["medium length message that gets", "broken in half"],
)
exact_length_case = ("first secon third", 5, ["first", "secon", "third"])


@pytest.mark.parametrize(
    "input,max_chunk_length,output_array",
    [
        single_line_case,
        medium_len_case,
        exact_length_case,
    ],
)
def test_chunk_message(input: str, max_chunk_length: int, output_array: List[str]):
    res = []

    for line in util.chunk_message(input, max_chunk_length):
        res.append(line)

    assert res == output_array


@pytest.mark.parametrize(
    "input,expect",
    [
        (None, None),
        ("1h", None),
        ("100hour", None),
        ("15sec", timedelta(seconds=15)),
        ("55min", timedelta(minutes=55)),
        ("1hour", timedelta(hours=1)),
        ("3day", timedelta(days=3)),
    ],
)
def test_calculate_time_delta(input, expect):
    res = util.calculate_time_delta(input)

    if expect is None:
        assert res is None
    else:
        assert res == expect


@pytest.mark.parametrize(
    "input_roles,role,expected_result",
    [
        ([], constants.Role.VERIFIED, False),
        ([constants.Role.MOD], constants.Role.VERIFIED, False),
        ([constants.Role.VERIFIED], constants.Role.VERIFIED, True),
        ([constants.Role.MOD, constants.Role.VERIFIED], constants.Role.VERIFIED, True),
    ],
)
def test_user_has_role(
    input_roles: List[constants.Role],
    role: constants.Role,
    expected_result: bool,
    create_member,
):
    mocked_member = create_member(roles=input_roles)

    res = util.user_has_role(mocked_member, role)

    assert res == expected_result


@pytest.mark.asyncio
async def test_add_and_remove_role(create_member):
    role_to_add = constants.Role.EXILED
    role_to_remove = constants.Role.VERIFIED

    mocked_member = create_member(roles=[constants.Role.VERIFIED])

    await util.add_and_remove_role(mocked_member, role_to_add, role_to_remove)

    mocked_member.add_roles.assert_called_once()
    mocked_member.remove_roles.assert_called_once()

    added_role = mocked_member.add_roles.call_args[0][0]
    assert added_role.name == role_to_add.value

    removed_role = mocked_member.remove_roles.call_args[0][0]
    assert removed_role.name == role_to_remove.value


def test_get_or_create_user_existing_user(mocker: MockerFixture, create_db_user):
    """Test that existing users are returned without creating new records"""
    discord_user_id = 12345
    existing_user = create_db_user(user_id=1, discord_user_id=str(discord_user_id))
    mock_logger = mocker.Mock()

    # Mock users_database module functions
    mock_get_user = mocker.patch("moddingway.database.users_database.get_user")
    mock_add_user = mocker.patch("moddingway.database.users_database.add_user")
    mock_get_user.return_value = existing_user

    result = util.get_or_create_user(discord_user_id, mock_logger)

    # Verify get_user was called
    mock_get_user.assert_called_once_with(discord_user_id)
    # Verify add_user was NOT called
    mock_add_user.assert_not_called()
    # Verify the existing user was returned
    assert result == existing_user


def test_get_or_create_user_new_user(mocker: MockerFixture, create_db_user):
    """Test that missing users are created"""
    discord_user_id = 67890
    new_user = create_db_user(user_id=2, discord_user_id=str(discord_user_id))
    mock_logger = mocker.Mock()

    # Mock users_database module functions
    mock_get_user = mocker.patch("moddingway.database.users_database.get_user")
    mock_add_user = mocker.patch("moddingway.database.users_database.add_user")
    mock_get_user.return_value = None  # User doesn't exist
    mock_add_user.return_value = new_user

    result = util.get_or_create_user(discord_user_id, mock_logger)

    # Verify get_user was called
    mock_get_user.assert_called_once_with(discord_user_id)
    # Verify add_user WAS called
    mock_add_user.assert_called_once_with(discord_user_id)
    # Verify the new user was returned
    assert result == new_user


def test_get_or_create_user_with_logging(
    mocker: MockerFixture, create_db_user, create_embed
):
    """Test that logging with embed works correctly"""
    discord_user_id = 11111
    new_user = create_db_user(user_id=3, discord_user_id=str(discord_user_id))
    mock_logger = mocker.Mock()
    mock_embed = create_embed()

    # Mock users_database module functions
    mock_get_user = mocker.patch("moddingway.database.users_database.get_user")
    mock_add_user = mocker.patch("moddingway.database.users_database.add_user")
    mock_get_user.return_value = None  # User doesn't exist
    mock_add_user.return_value = new_user

    # Mock log_info_and_embed
    mock_log_info = mocker.patch("moddingway.util.log_info_and_embed")

    result = util.get_or_create_user(discord_user_id, mock_logger, mock_embed)

    # Verify logging was called with correct arguments
    mock_log_info.assert_called_once_with(
        mock_embed,
        mock_logger,
        "User not found in database, creating new record",
    )
    # Verify user was created
    mock_add_user.assert_called_once_with(discord_user_id)
    assert result == new_user


def test_get_or_create_user_with_logger_only(mocker: MockerFixture, create_db_user):
    """Test that logger-only logging works correctly"""
    discord_user_id = 22222
    new_user = create_db_user(user_id=4, discord_user_id=str(discord_user_id))
    mock_logger = mocker.Mock()

    # Mock users_database module functions
    mock_get_user = mocker.patch("moddingway.database.users_database.get_user")
    mock_add_user = mocker.patch("moddingway.database.users_database.add_user")
    mock_get_user.return_value = None  # User doesn't exist
    mock_add_user.return_value = new_user

    result = util.get_or_create_user(discord_user_id, mock_logger)

    # Verify logger.info was called with correct message
    mock_logger.info.assert_called_once_with(
        f"Creating new user record for Discord ID {discord_user_id}"
    )
    # Verify user was created
    mock_add_user.assert_called_once_with(discord_user_id)
    assert result == new_user


def test_get_or_create_user_without_embed(mocker: MockerFixture, create_db_user):
    """Test that function works without embed (logger only)"""
    discord_user_id = 33333
    new_user = create_db_user(user_id=5, discord_user_id=str(discord_user_id))
    mock_logger = mocker.Mock()

    # Mock users_database module functions
    mock_get_user = mocker.patch("moddingway.database.users_database.get_user")
    mock_add_user = mocker.patch("moddingway.database.users_database.add_user")
    mock_get_user.return_value = None  # User doesn't exist
    mock_add_user.return_value = new_user

    result = util.get_or_create_user(discord_user_id, mock_logger)

    # Verify logger.info was called (no embed path)
    mock_logger.info.assert_called_once()
    # Verify user was created
    mock_add_user.assert_called_once_with(discord_user_id)
    # Verify the new user was returned
    assert result == new_user
