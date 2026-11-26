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


@pytest.mark.parametrize(
    "message,from_index,max_chunk_length,expected",
    [
        # splits after the last newline inside the window, not one before it or past it
        ("a\nb\ncdefg", 0, 6, 4),
        ("x\nabcdefgh", 2, 4, 6),
        ("abcdef\ngh", 0, 4, 4),
        # splits after the last space inside the window
        ("ab cd efgh", 0, 7, 6),
        # otherwise cuts at the window edge
        ("abcdefgh", 0, 4, 4),
    ],
)
def test_split_chunks(message, from_index, max_chunk_length, expected):
    assert util._split_chunks(message, from_index, max_chunk_length) == expected


class _CodedError(Exception):
    def __init__(self, code):
        super().__init__(f"coded error {code}")
        self.code = code


@pytest.mark.asyncio
async def test_send_dm__sends_message(create_member):
    member = create_member(id=42)
    embed = discord.Embed()

    await util.send_dm(embed, member, "hello", context="Exile")

    member.create_dm.return_value.send.assert_awaited_once_with(content="hello")
    assert embed.fields == []


@pytest.mark.asyncio
async def test_send_dm__known_error_code_logs_template(create_member, embed_fields):
    member = create_member(id=42)
    member.create_dm.side_effect = _CodedError(50007)
    embed = discord.Embed()

    await util.send_dm(embed, member, "hello", context="Exile")

    assert embed_fields(embed) == [
        (
            "DM Status",
            constants.ERROR_MESSAGES[50007].format(user=42, context="Exile"),
        )
    ]


@pytest.mark.parametrize(
    "error", [_CodedError(12345), _CodedError("50007"), Exception("boom")]
)
@pytest.mark.asyncio
async def test_send_dm__other_errors_log_fallback(create_member, error, embed_fields):
    member = create_member(id=42)
    member.create_dm.side_effect = error
    embed = discord.Embed()

    await util.send_dm(embed, member, "hello", context="Exile")

    assert embed_fields(embed) == [
        ("DM Status", f"Failed to send DM to <@42> for Exile, {error!s}")
    ]


def test_chunk_message__default_length_is_2000():
    assert list(util.chunk_message("a" * 2001)) == ["a" * 2000, "a"]


@pytest.mark.parametrize(
    "missing,expected",
    [
        (constants.Role.EXILED, "Role Exiled not found in server"),
        (constants.Role.VERIFIED, "Role Verified not found in server"),
    ],
)
@pytest.mark.asyncio
async def test_add_and_remove_role__missing_role_raises(
    create_member, create_role, missing, expected
):
    member = create_member()
    member.guild.roles = [
        role for role in member.guild.roles if role.name != missing.value
    ]

    with pytest.raises(Exception, match=f"^{expected}$"):
        await util.add_and_remove_role(
            member,
            role_to_add=constants.Role.EXILED,
            role_to_remove=constants.Role.VERIFIED,
        )

    member.remove_roles.assert_not_called()
    member.add_roles.assert_not_called()


@pytest.mark.asyncio
async def test_is_user_moderator__non_member_is_not_moderator(mocker):
    interaction = mocker.Mock(user=mocker.Mock(spec=discord.User))

    assert await util.is_user_moderator(interaction) is False


def test_log_info_and_add_field__logs_name_and_value(mocker, embed_fields):
    embed = discord.Embed()
    logger = mocker.Mock()

    util.log_info_and_add_field(embed, logger, "Result", "done")

    logger.info.assert_called_once_with("Result: done")
    assert embed_fields(embed) == [("Result", "done")]
