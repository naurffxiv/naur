from datetime import UTC, datetime

import discord
import pytest
from fakes import embed_fields
from pytest_mock.plugin import MockerFixture

from moddingway import constants
from moddingway.constants import ExileStatus
from moddingway.database.models import Exile
from moddingway.services import exile_service


@pytest.mark.parametrize(
    "status,deleted,expected",
    [
        (ExileStatus.TIMED_EXILED, None, "Selected exile is active, cannot be deleted"),
        (ExileStatus.UNEXILED, True, "Exile 7 successfully deleted"),
        (ExileStatus.UNEXILED, False, "Error with exile 7 deletion"),
        (None, None, "Exile 7 not found"),
    ],
)
@pytest.mark.asyncio
async def test_delete_exile_by_id(mocker: MockerFixture, status, deleted, expected):
    mocker.patch(
        "moddingway.database.exiles_database.get_exile_status", return_value=status
    )
    delete = mocker.patch(
        "moddingway.database.exiles_database.delete_exile", return_value=deleted
    )
    embed = discord.Embed()

    assert await exile_service.delete_exile_by_id(embed, 7) == expected

    assert delete.called == (status == ExileStatus.UNEXILED)
    assert embed_fields(embed) == [("Result", expected)]


@pytest.fixture
def unexile_mocks(mocker: MockerFixture, create_db_user):
    return {
        "get_user": mocker.patch(
            "moddingway.database.users_database.get_user",
            return_value=create_db_user(user_id=1),
        ),
        "active": mocker.patch(
            "moddingway.database.exiles_database.get_user_active_exile",
            return_value=mocker.Mock(exile_id=9),
        ),
        "update_status": mocker.patch(
            "moddingway.database.exiles_database.update_exile_status"
        ),
        "get_sticky": mocker.patch(
            "moddingway.database.roles_database.get_sticky_roles", return_value=[]
        ),
        "remove_sticky": mocker.patch(
            "moddingway.database.roles_database.remove_sticky_roles"
        ),
    }


@pytest.mark.asyncio
async def test_unexile_user_restores_verified_and_closes_the_exile(
    create_member, unexile_mocks
):
    member = create_member(roles=[constants.Role.EXILED], id=12345)
    embed = discord.Embed()

    assert await exile_service.unexile_user(embed, member) is None

    member.remove_roles.assert_awaited_once()
    assert member.remove_roles.await_args.args[0].name == constants.Role.EXILED.value
    assert member.add_roles.await_args.args[0].name == constants.Role.VERIFIED.value
    member.create_dm.return_value.send.assert_awaited_once_with(
        content="You have been un-exiled from NA Ultimate Raiding - FFXIV."
    )
    unexile_mocks["update_status"].assert_called_once_with(9, ExileStatus.UNEXILED)
    assert embed.footer.text == "Exile ID: 9"
    assert embed_fields(embed)[-1] == ("Result", "<@12345> was successfully unexiled")


@pytest.mark.asyncio
async def test_unexile_user_restores_sticky_roles(
    create_member, unexile_mocks, mocker: MockerFixture
):
    member = create_member(roles=[constants.Role.EXILED], id=12345)
    sticky = mocker.Mock(id=500)
    member.guild.get_role = mocker.Mock(side_effect={500: sticky}.get)
    unexile_mocks["get_sticky"].return_value = ["500", "600"]

    await exile_service.unexile_user(discord.Embed(), member)

    member.add_roles.assert_any_await(sticky)
    unexile_mocks["remove_sticky"].assert_called_once_with(1)


@pytest.mark.asyncio
async def test_unexile_user_without_active_exile_skips_the_database(
    create_member, unexile_mocks
):
    unexile_mocks["active"].return_value = None
    member = create_member(roles=[constants.Role.EXILED], id=12345)
    embed = discord.Embed()

    await exile_service.unexile_user(embed, member)

    unexile_mocks["update_status"].assert_not_called()
    assert embed.footer.text is None


@pytest.mark.asyncio
async def test_unexile_user_rejects_members_who_are_not_exiled(
    create_member, unexile_mocks
):
    member = create_member(roles=[constants.Role.VERIFIED], id=12345)

    result = await exile_service.unexile_user(discord.Embed(), member)

    assert result == "User is not currently exiled, no action will be taken"
    member.add_roles.assert_not_awaited()


@pytest.mark.asyncio
async def test_unexile_user_reports_missing_database_record(
    create_member, unexile_mocks
):
    unexile_mocks["get_user"].return_value = None
    member = create_member(roles=[constants.Role.EXILED], id=12345)

    result = await exile_service.unexile_user(discord.Embed(), member)

    assert (
        result == "User has been unexiled, but no user record was found in the database"
    )
    member.add_roles.assert_awaited_once()


START = datetime(2025, 9, 1, tzinfo=UTC)
END = datetime(2025, 9, 2, tzinfo=UTC)


@pytest.mark.asyncio
async def test_get_user_exiles(mocker: MockerFixture, create_db_user):
    mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=create_db_user(user_id=1),
    )
    mocker.patch(
        "moddingway.database.exiles_database.get_user_exiles",
        return_value=[
            Exile(3, 1, "12345", "spam", ExileStatus.TIMED_EXILED, START, END)
        ],
    )

    result = await exile_service.get_user_exiles(mocker.Mock(id=12345))

    assert result == (
        "Exiles found for <@12345>:"
        "\n* ID: 3 | START DATE: <t:1756684800:F> | END DATE: <t:1756771200:F> | TYPE: TIMED_EXILED | REASON: spam"
    )


@pytest.mark.asyncio
async def test_get_user_exiles_when_missing(mocker: MockerFixture, create_db_user):
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user", return_value=None
    )
    mocker.patch("moddingway.database.exiles_database.get_user_exiles", return_value=[])

    assert (
        await exile_service.get_user_exiles(mocker.Mock(id=1))
        == "User not found in database"
    )
    get_user.return_value = create_db_user(user_id=1)
    assert (
        await exile_service.get_user_exiles(mocker.Mock(id=1))
        == "No exiles found for user"
    )


@pytest.mark.asyncio
async def test_get_active_exiles(mocker: MockerFixture):
    get_all = mocker.patch(
        "moddingway.database.exiles_database.get_all_active_exiles",
        return_value=[
            Exile(3, 1, "12345", "spam", ExileStatus.TIMED_EXILED, START, END)
        ],
    )

    assert await exile_service.get_active_exiles() == (
        "Active exiles found for"
        "\n* ID: 3 | USER: <@12345> | START DATE: <t:1756684800:F> | END DATE: <t:1756771200:F> | TYPE: TIMED_EXILED | REASON: spam"
    )
    get_all.return_value = []
    assert await exile_service.get_active_exiles() == "No active exiles found"


@pytest.mark.parametrize(
    "duration,expected",
    [
        ("1sec", "1 second"),
        ("30sec", "30 seconds"),
        ("1min", "1 minute"),
        ("2hour", "2 hours"),
        ("1day", "1 day"),
        ("100day", None),
        ("", None),
        (None, None),
    ],
)
def test_format_time_string(duration, expected):
    assert exile_service.format_time_string(duration) == expected
