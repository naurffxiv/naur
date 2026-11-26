import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway.constants import MAX_SEND_MESSAGE_LENGTH
from moddingway.services.send_message_service import send_message_to_channel


@pytest.mark.asyncio
async def test_send_message_to_channel__success(mocker: MockerFixture):
    # Arrange
    logging_embed = mocker.Mock()
    channel = mocker.Mock(mention="#announcements")
    channel.send = mocker.AsyncMock()

    # Act
    result = await send_message_to_channel(logging_embed, channel, "Hello world")

    # Assert
    assert result == "Successfully sent message to #announcements"
    channel.send.assert_called_once_with("Hello world")
    logging_embed.add_field.assert_called_once_with(
        name="Result",
        value="Successfully sent message to #announcements",
        inline=False,
    )


@pytest.mark.asyncio
async def test_send_message_to_channel__message_too_long(mocker: MockerFixture):
    # Arrange
    logging_embed = mocker.Mock()
    channel = mocker.Mock(mention="#mod-chat")
    channel.send = mocker.AsyncMock()

    message = "x" * (MAX_SEND_MESSAGE_LENGTH + 1)

    # Act
    result = await send_message_to_channel(logging_embed, channel, message)

    # Assert
    assert result == f"Message must be {MAX_SEND_MESSAGE_LENGTH} characters or less."
    channel.send.assert_not_called()
    logging_embed.add_field.assert_called_once_with(
        name="Error",
        value=f"Message must be {MAX_SEND_MESSAGE_LENGTH} characters or less.",
        inline=False,
    )


@pytest.mark.asyncio
async def test_send_message_to_channel__unexpected_send_error(mocker: MockerFixture):
    # Arrange
    logging_embed = mocker.Mock()
    channel = mocker.Mock(mention="#announcements")
    channel.send = mocker.AsyncMock(side_effect=Exception("send failed"))

    # Act
    result = await send_message_to_channel(logging_embed, channel, "Hello world")

    # Assert
    assert result == "Unable to post in #announcements: an unexpected error occurred."
    logging_embed.add_field.assert_called_once_with(
        name="Error",
        value="Unable to post in #announcements: an unexpected error occurred.",
        inline=False,
    )


@pytest.mark.asyncio
async def test_send_message_to_channel__message_at_limit_is_sent(
    mocker: MockerFixture,
):
    channel = mocker.Mock(mention="#announcements")
    channel.send = mocker.AsyncMock()
    message = "x" * MAX_SEND_MESSAGE_LENGTH

    result = await send_message_to_channel(mocker.Mock(), channel, message)

    assert result == "Successfully sent message to #announcements"
    channel.send.assert_called_once_with(message)


def _http_response(mocker: MockerFixture, status: int):
    return mocker.Mock(status=status, reason="reason")


@pytest.mark.parametrize(
    "error_type,status,expected",
    [
        (
            discord.Forbidden,
            403,
            "Unable to post in #announcements: the bot is missing permissions.",
        ),
        (
            discord.HTTPException,
            500,
            "Unable to post in #announcements: 500 reason (error code: 0): boom",
        ),
    ],
)
@pytest.mark.asyncio
async def test_send_message_to_channel__discord_errors(
    mocker: MockerFixture, error_type, status, expected, embed_fields
):
    channel = mocker.Mock(mention="#announcements")
    channel.send = mocker.AsyncMock(
        side_effect=error_type(_http_response(mocker, status), "boom")
    )
    logging_embed = discord.Embed()

    result = await send_message_to_channel(logging_embed, channel, "Hello world")

    assert result == expected
    assert embed_fields(logging_embed) == [("Error", expected)]
