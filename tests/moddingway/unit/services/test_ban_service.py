import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway.constants import MAX_BAN_REASON_LENGTH
from moddingway.services import ban_service


@pytest.fixture
def member(create_member, mocker: MockerFixture):
    m = create_member(id=12345)
    m.mention = "<@12345>"
    m.display_name = "Alex"
    m.ban = mocker.AsyncMock()
    return m


@pytest.mark.asyncio
async def test_ban_member_dms_then_bans(member):
    ok, message = await ban_service.ban_member(
        discord.Embed(), member, "spam", delete_messages=True
    )

    assert (ok, message) == (True, None)
    member.ban.assert_awaited_once_with(reason="spam", delete_message_seconds=604800)
    dm = member.create_dm.return_value.send.await_args.kwargs["content"]
    assert dm.startswith(
        "Hello Alex,\n\nYou are being informed that you have been **banned**"
    )
    assert "> spam\n\n" in dm


@pytest.mark.asyncio
async def test_ban_member_can_keep_messages(member):
    await ban_service.ban_member(discord.Embed(), member, "spam", delete_messages=False)

    member.ban.assert_awaited_once_with(reason="spam", delete_message_seconds=0)


@pytest.mark.asyncio
async def test_ban_accepts_reason_at_the_limit_and_rejects_longer(member):
    ok, _ = await ban_service.ban_member(
        discord.Embed(), member, "x" * MAX_BAN_REASON_LENGTH, False
    )
    assert ok

    ok, message = await ban_service.ban_member(
        discord.Embed(), member, "x" * (MAX_BAN_REASON_LENGTH + 1), False
    )
    assert not ok
    assert message == (
        f"Unable to ban <@12345>: reason is too long (above {MAX_BAN_REASON_LENGTH} characters). "
        "Please shorten the ban reason."
    )
    member.ban.assert_awaited_once()


@pytest.mark.asyncio
async def test_ban_member_reports_discord_failures(member, mocker: MockerFixture):
    member.ban.side_effect = discord.HTTPException(
        mocker.Mock(status=500, reason="x"), "boom"
    )

    ok, message = await ban_service.ban_member(discord.Embed(), member, "spam", False)

    assert not ok
    assert (
        message
        == "Failed to ban <@12345>. Please try again or use Discord's built-in tools."
    )


@pytest.mark.asyncio
async def test_ban_user_bans_through_the_guild(member, mocker: MockerFixture):
    interaction = mocker.Mock(guild=mocker.Mock(ban=mocker.AsyncMock()))

    ok, _ = await ban_service.ban_user(
        discord.Embed(), interaction, member, "spam", True
    )

    assert ok
    interaction.guild.ban.assert_awaited_once_with(
        member, reason="spam", delete_message_seconds=604800
    )


@pytest.mark.asyncio
async def test_ban_user_needs_a_guild(member, mocker: MockerFixture):
    ok, message = await ban_service.ban_user(
        discord.Embed(), mocker.Mock(guild=None), member, "spam", True
    )

    assert not ok
    assert (
        message
        == "Unable to perform this action because the guild information is missing."
    )
