from datetime import timedelta

import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import constants, util
from moddingway.constants import UserRole
from moddingway.database.models import Exile, User
from moddingway.services import exile_service


@pytest.mark.asyncio
async def test_exile_user__unverified(
    mocker: MockerFixture, create_member, embed_fields
):
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
    # Act
    logging_embed = discord.Embed()
    res = await exile_service.exile_user(
        logging_embed,
        mocked_member,
        timedelta(days=1),
        "test_exile_user__unverified",
    )

    # Assert
    assert res is not None
    assert res == "User is not currently verified, no action will be taken"
    assert embed_fields(logging_embed) == [("Error", res)]


@pytest.mark.asyncio
async def test_exile_user__verified_existing_user_dm_failed(
    mocker: MockerFixture, create_member
):
    # Arrange
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

    mocked_member = create_member(roles=[constants.Role.VERIFIED], allows_dms=False)
    mocked_logging_embed = mocker.Mock()

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
    # TODO check exile create call to confirm data is correct

    assert any(
        call[1].get("name", "") == "DM Status"
        for call in mocked_logging_embed.add_field.call_args_list
    )


def _active_exile(start, end):
    return Exile(
        exile_id=7,
        user_id=1,
        discord_id=12345,
        reason="earlier",
        exile_status=constants.ExileStatus.TIMED_EXILED.value,
        start_timestamp=start,
        end_timestamp=end,
    )


@pytest.fixture
def exile_mocks(mocker: MockerFixture, create_db_user):
    return {
        "get_user": mocker.patch(
            "moddingway.database.users_database.get_user",
            return_value=create_db_user(user_id=1),
        ),
        "add_user": mocker.patch(
            "moddingway.database.users_database.add_user",
            return_value=create_db_user(user_id=1),
        ),
        "get_user_active_exile": mocker.patch(
            "moddingway.database.exiles_database.get_user_active_exile",
            return_value=None,
        ),
        "add_exile": mocker.patch(
            "moddingway.database.exiles_database.add_exile", return_value=4001
        ),
        "update_exile_end": mocker.patch(
            "moddingway.database.exiles_database.update_exile_end"
        ),
        "add_sticky_roles": mocker.patch(
            "moddingway.database.roles_database.add_sticky_roles"
        ),
    }


@pytest.mark.asyncio
async def test_exile_user__extends_active_exile(
    create_member, exile_mocks, default_now, embed_fields
):
    end = default_now + timedelta(days=2)
    exile_mocks["get_user_active_exile"].return_value = _active_exile(default_now, end)
    member = create_member(roles=[constants.Role.EXILED], id=12345)
    embed = discord.Embed()

    res = await exile_service.exile_user(embed, member, timedelta(days=1), "again")

    new_end = end + timedelta(days=1)
    assert res == "User exile extended"
    exile_mocks["get_user_active_exile"].assert_called_once_with(1)
    exile_mocks["update_exile_end"].assert_called_once_with(7, new_end)
    exile_mocks["add_exile"].assert_not_called()
    assert embed_fields(embed) == [
        ("Result", f"Exile extended until <t:{util.timestamp_to_epoch(new_end)}>")
    ]


@pytest.mark.asyncio
async def test_exile_user__verified_user_with_active_exile_is_not_extended(
    create_member, exile_mocks, default_now
):
    exile_mocks["get_user_active_exile"].return_value = _active_exile(
        default_now, default_now
    )
    member = create_member(roles=[constants.Role.VERIFIED], id=12345)

    res = await exile_service.exile_user(
        discord.Embed(), member, timedelta(days=1), "reason"
    )

    assert res is None
    exile_mocks["update_exile_end"].assert_not_called()
    exile_mocks["add_exile"].assert_called_once()


@pytest.mark.asyncio
async def test_exile_user__creates_exile_and_messages_user(
    create_member, exile_mocks, default_now, embed_fields
):
    member = create_member(roles=[constants.Role.VERIFIED], id=12345)
    embed = discord.Embed()
    end = default_now + timedelta(days=1)
    epoch = util.timestamp_to_epoch(end)

    res = await exile_service.exile_user(embed, member, timedelta(days=1), "spam")

    assert res is None
    exile_mocks["get_user"].assert_called_once_with(12345)
    exile_mocks["add_user"].assert_not_called()
    (exile,), _ = exile_mocks["add_exile"].call_args
    assert vars(exile) == {
        "exile_id": None,
        "user_id": 1,
        "discord_id": 12345,
        "reason": "spam",
        "exile_status": constants.ExileStatus.TIMED_EXILED.value,
        "start_timestamp": default_now,
        "end_timestamp": end,
    }
    assert embed.footer.text == "Exile ID: 4001"
    member.remove_roles.assert_awaited_once()
    member.add_roles.assert_awaited_once()
    member.create_dm.return_value.send.assert_awaited_once_with(
        content=(
            "You are being exiled from NA Ultimate Raiding - FFXIV.\n"
            f"**Reason:** spam\nExile expiration: <t:{epoch}:R>"
        )
    )
    assert embed_fields(embed) == [
        ("Expiration", f"<t:{epoch}:R>"),
        ("Result", "<@12345> was successfully exiled"),
    ]
    exile_mocks["add_sticky_roles"].assert_not_called()


@pytest.mark.asyncio
async def test_exile_user__creates_missing_database_user(create_member, exile_mocks):
    exile_mocks["get_user"].return_value = None
    member = create_member(roles=[constants.Role.VERIFIED], id=12345)
    embed = discord.Embed()

    await exile_service.exile_user(embed, member, timedelta(days=1), "spam")

    exile_mocks["add_user"].assert_called_once_with(12345)
    assert embed.description == "\nUser not found in database, creating new record"
    (exile,), _ = exile_mocks["add_exile"].call_args
    assert exile.user_id == 1


@pytest.mark.asyncio
async def test_exile_user__strips_and_saves_sticky_roles(
    mocker: MockerFixture, create_member, exile_mocks, monkeypatch
):
    monkeypatch.setattr(exile_service.settings, "sticky_roles", [500, 600])
    member = create_member(roles=[constants.Role.VERIFIED], id=12345)
    sticky = mocker.Mock(id=500)
    missing = mocker.Mock(id=600)
    other = mocker.Mock(id=700)
    member.roles = [*member.roles, sticky, missing, other]
    guild_roles = {500: sticky}
    member.guild.get_role = mocker.Mock(side_effect=guild_roles.get)

    await exile_service.exile_user(discord.Embed(), member, timedelta(days=1), "spam")

    member.remove_roles.assert_any_await(sticky)
    exile_mocks["add_sticky_roles"].assert_called_once_with(1, [500])


@pytest.mark.asyncio
async def test_exile_user__logs_sticky_role_failure(
    mocker: MockerFixture, create_member, exile_mocks, monkeypatch, embed_fields
):
    monkeypatch.setattr(exile_service.settings, "sticky_roles", [500])
    member = create_member(roles=[constants.Role.VERIFIED], id=12345)
    member.roles = [*member.roles, mocker.Mock(id=500)]
    member.guild.get_role = mocker.Mock(side_effect=RuntimeError("boom"))
    embed = discord.Embed()

    await exile_service.exile_user(embed, member, timedelta(days=1), "spam")

    assert (
        "Error",
        "Sticky roles could not be completely removed, boom",
    ) in embed_fields(embed)


@pytest.mark.asyncio
async def test_exile_user__dm_failure_names_exile_context(
    create_member, exile_mocks, embed_fields
):
    member = create_member(roles=[constants.Role.VERIFIED], id=12345, allows_dms=False)
    embed = discord.Embed()

    await exile_service.exile_user(embed, member, timedelta(days=1), "spam")

    assert (
        "DM Status",
        "Failed to send DM to <@12345> for Exile, Cannot create DM",
    ) in embed_fields(embed)
