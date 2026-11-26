from datetime import UTC, datetime, timedelta

import discord
import pytest
from discord.ext import commands
from fakes import embed_fields
from pytest_mock.plugin import MockerFixture

from moddingway import constants
from moddingway.events.member_events import register_events


@pytest.fixture
def bot():
    b = commands.Bot(command_prefix="!", intents=discord.Intents.none())
    register_events(b)
    return b


@pytest.fixture
def log_channel(mocker: MockerFixture):
    channel = mocker.Mock(spec=discord.TextChannel)
    channel.send = mocker.AsyncMock()
    return channel


def _guild(mocker, log_channel, roles=()):
    guild = mocker.Mock()
    guild.get_channel.return_value = log_channel
    guild.roles = list(roles)
    return guild


def _role(mocker, name, id=1):
    role = mocker.Mock(id=id)
    role.name = name
    return role


def _fields(channel):
    return embed_fields(channel.send.await_args.kwargs["embed"])


@pytest.mark.asyncio
async def test_member_join_assigns_non_verified_and_logs_account_age(
    bot, log_channel, mocker: MockerFixture
):
    non_verified = _role(mocker, constants.Role.NON_VERIFIED.value, id=55)
    member = mocker.Mock(
        id=12345,
        display_name="new",
        display_avatar=mocker.Mock(url="u"),
        guild=_guild(mocker, log_channel, [non_verified]),
        add_roles=mocker.AsyncMock(),
        created_at=datetime.now(UTC) - timedelta(days=10, hours=1),
    )

    await bot.on_member_join(member)

    member.add_roles.assert_awaited_once_with(non_verified)
    embed = log_channel.send.await_args.kwargs["embed"]
    assert embed.description == "<@12345> joined the server"
    assert embed.footer.text == "Member Joined"
    fields = _fields(log_channel)
    assert fields[0] == ("Result", "Successfully assigned <@&55> role")
    assert fields[1][0] == "Account Age"
    assert fields[1][1].startswith("10 days (Created: ")


@pytest.mark.asyncio
async def test_member_join_logs_role_errors(bot, log_channel, mocker: MockerFixture):
    member = mocker.Mock(
        id=1,
        display_name="new",
        display_avatar=mocker.Mock(url="u"),
        guild=_guild(mocker, log_channel),
    )

    await bot.on_member_join(member)

    assert _fields(log_channel) == [
        ("Error", "Role 'Non-Verified' not found in the server.")
    ]


@pytest.mark.asyncio
async def test_member_join_without_log_channel_still_assigns_role(
    bot, mocker: MockerFixture
):
    non_verified = _role(mocker, constants.Role.NON_VERIFIED.value)
    member = mocker.Mock(
        guild=_guild(mocker, None, [non_verified]), add_roles=mocker.AsyncMock()
    )

    await bot.on_member_join(member)

    member.add_roles.assert_awaited_once_with(non_verified)


@pytest.mark.parametrize(
    "event,banned,verb",
    [("on_member_ban", True, "banned"), ("on_member_unban", False, "unbanned")],
)
@pytest.mark.parametrize("existing", [True, False])
@pytest.mark.asyncio
async def test_ban_events_update_the_user_record(
    bot,
    log_channel,
    mocker: MockerFixture,
    create_db_user,
    event,
    banned,
    verb,
    existing,
):
    db_user = create_db_user(user_id=1)
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=db_user if existing else None,
    )
    add_user = mocker.patch(
        "moddingway.database.users_database.add_user", return_value=db_user
    )
    update = mocker.patch("moddingway.database.users_database.update_user")
    user = mocker.Mock(id=12345, mention="<@12345>")

    await getattr(bot, event)(_guild(mocker, log_channel), user)

    get_user.assert_called_once_with(12345)
    assert add_user.called is not existing
    assert db_user.is_banned is banned
    update.assert_called_once_with(db_user)
    embed = log_channel.send.await_args.kwargs["embed"]
    assert embed.description == f"<@12345> was {verb}"
    assert _fields(log_channel)[-1] == ("Result", f"Successfully {verb} <@12345>")


@pytest.mark.asyncio
async def test_ban_event_without_log_channel_still_records_ban(
    bot, mocker: MockerFixture, create_db_user
):
    db_user = create_db_user(user_id=1)
    mocker.patch("moddingway.database.users_database.get_user", return_value=db_user)
    update = mocker.patch("moddingway.database.users_database.update_user")

    await bot.on_member_ban(_guild(mocker, None), mocker.Mock(id=1, mention="<@1>"))

    update.assert_called_once_with(db_user)
