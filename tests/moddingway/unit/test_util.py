from datetime import timedelta

import discord
import pytest

from moddingway import constants, util

single_line_case = ("short message", 100, ["short message"])
medium_len_case = (
    "medium length message that gets broken in half",
    35,
    ["medium length message that gets", "broken in half"],
)
exact_length_case = ("first second third", 6, ["first", "second", "third"])


@pytest.mark.parametrize(
    "input,max_chunk_length,output_array",
    [
        single_line_case,
        medium_len_case,
        exact_length_case,
    ],
)
def test_chunk_message(input: str, max_chunk_length: int, output_array: list[str]):
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
    input_roles: list[constants.Role],
    role: constants.Role,
    expected_result: bool,
    create_member,
):
    mocked_member = create_member(roles=input_roles)

    res = util.user_has_role(mocked_member, role)

    assert res == expected_result


@pytest.mark.asyncio
async def test_add_and_remove_role__swaps_in_one_edit(
    mocker, create_member, naur_guild
):
    verified, exiled = (
        util.get_guild_role(naur_guild, role)
        for role in (constants.Role.VERIFIED, constants.Role.EXILED)
    )
    mocked_member = create_member(roles=[constants.Role.MOD])
    mocked_member.edit = mocker.AsyncMock()
    other_role = mocked_member.roles[0]
    mocked_member.roles.append(verified)

    await util.add_and_remove_role(
        mocked_member, constants.Role.EXILED, constants.Role.VERIFIED
    )

    mocked_member.edit.assert_awaited_once_with(roles=[other_role, exiled])
    mocked_member.add_roles.assert_not_called()
    mocked_member.remove_roles.assert_not_called()


@pytest.mark.parametrize(
    "duration,expected",
    [
        (None, None),
        ("1h", None),
        ("5sec", (5, "second")),
        ("12min", (12, "minute")),
        ("05hour", (5, "hour")),
        ("3day", (3, "day")),
    ],
)
def test_parse_duration(duration, expected):
    assert util.parse_duration(duration) == expected


@pytest.mark.parametrize(
    "input_roles,expected_result",
    [
        ([constants.Role.MOD], True),
        ([constants.Role.ADMIN], True),
        ([constants.Role.MOD, constants.Role.ADMIN], True),
        ([constants.Role.VERIFIED], False),
        ([], False),
    ],
)
@pytest.mark.asyncio
async def test_is_user_moderator(
    mocker,
    create_member,
    input_roles: list[constants.Role],
    expected_result: bool,
):
    mocked_member = create_member(roles=input_roles)

    mocked_interaction = mocker.Mock(spec=discord.Interaction)
    mocked_interaction.user = mocked_member

    if expected_result:
        result = await util.is_user_moderator(mocked_interaction)
        assert result is True
    else:
        with pytest.raises(discord.app_commands.MissingAnyRole):
            await util.is_user_moderator(mocked_interaction)
