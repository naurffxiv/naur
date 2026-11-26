import discord
import pytest
from fakes import make_user
from pytest_mock.plugin import MockerFixture

from moddingway import constants, util
from moddingway.constants import UserRole
from moddingway.database.models import PendingExile


@pytest.mark.asyncio
async def test_find_and_assign_role(create_member):
    member = create_member()
    verified = next(
        r for r in member.guild.roles if r.name == constants.Role.VERIFIED.value
    )
    verified.id = 77

    assert await util.find_and_assign_role(member, constants.Role.VERIFIED) == (
        True,
        "Successfully assigned <@&77> role",
        verified,
    )
    member.add_roles.assert_awaited_once_with(verified)


@pytest.mark.asyncio
async def test_find_and_assign_role_failures(create_member, mocker: MockerFixture):
    member = create_member()
    member.display_name = "Alex"

    assert await util.find_and_assign_role(member, constants.Role.NON_VERIFIED) == (
        False,
        "Role 'Non-Verified' not found in the server.",
        None,
    )

    member.add_roles.side_effect = discord.Forbidden(
        mocker.Mock(status=403, reason="x"), "no"
    )
    assert await util.find_and_assign_role(member, constants.Role.VERIFIED) == (
        False,
        "Bot does not have permission to add roles to Alex",
        None,
    )

    member.add_roles.side_effect = RuntimeError("boom")
    assert await util.find_and_assign_role(member, constants.Role.VERIFIED) == (
        False,
        "Failed to assign role: boom",
        None,
    )


def test_get_log_channel(mocker: MockerFixture):
    bot = mocker.Mock()
    bot.get_channel.return_value = "channel"
    assert util.get_log_channel(bot) == "channel"
    bot.get_channel.assert_called_once_with(util.settings.logging_channel_id)

    bot.get_channel.return_value = None
    assert util.get_log_channel(bot) is None


@pytest.mark.parametrize(
    "role,mod,sysadmin",
    [
        (UserRole.USER, False, False),
        (UserRole.MOD, True, False),
        (UserRole.SYSADMIN, True, True),
    ],
)
def test_user_permissions(role, mod, sysadmin):
    user = make_user(role=role, temporary=2, permanent=7)

    assert user.has_mod_permissions() is mod
    assert user.has_sysadmin_permissions() is sysadmin
    assert (user.temporary_points, user.permanent_points, user.get_strike_points()) == (
        2,
        7,
        9,
    )


def test_pending_exile_coerces_ids_to_int():
    pending = PendingExile("3", "1", "12345", "end")

    assert (
        pending.exile_id,
        pending.user_id,
        pending.discord_id,
        pending.end_timestamp,
    ) == (3, 1, 12345, "end")
