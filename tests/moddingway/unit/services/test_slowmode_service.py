import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway.constants import MAX_SLOWMODE_INTERVAL
from moddingway.services import slowmode_service


def _channel(mocker: MockerFixture, current: int):
    return mocker.Mock(slowmode_delay=current, mention="#chat", edit=mocker.AsyncMock())


@pytest.mark.parametrize(
    "current,interval,expected,edited",
    [
        (0, 30, "Successfully set slowmode to 30 seconds in #chat", True),
        (30, 0, "Successfully turned off slowmode in #chat", True),
        (30, 30, "Slowmode is already set to 30 seconds in #chat", False),
        (0, 0, "Slowmode is already off in #chat", False),
        (
            0,
            MAX_SLOWMODE_INTERVAL,
            f"Successfully set slowmode to {MAX_SLOWMODE_INTERVAL} seconds in #chat",
            True,
        ),
        (
            0,
            MAX_SLOWMODE_INTERVAL + 1,
            f"Interval must be between 0 and {MAX_SLOWMODE_INTERVAL} seconds",
            False,
        ),
        (
            0,
            -1,
            f"Interval must be between 0 and {MAX_SLOWMODE_INTERVAL} seconds",
            False,
        ),
    ],
)
@pytest.mark.asyncio
async def test_edit_slowmode(
    mocker: MockerFixture, current, interval, expected, edited
):
    channel = _channel(mocker, current)
    embed = discord.Embed()

    assert await slowmode_service.edit_slowmode(embed, interval, channel) == expected

    expected_edits = [mocker.call(slowmode_delay=interval)] if edited else []
    assert channel.edit.await_args_list == expected_edits
    assert embed.fields[0].value == expected
