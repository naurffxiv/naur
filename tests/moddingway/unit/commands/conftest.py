import datetime

import discord
import pytest
from discord.ext import commands
from pytest_mock.plugin import MockerFixture


@pytest.fixture(autouse=True)
def mock_datetime_now():
    """Commands build embeds with real timestamps, which discord.py type-checks."""


@pytest.fixture
def bot():
    return commands.Bot(command_prefix="!", intents=discord.Intents.none())


@pytest.fixture
def make_interaction(mocker: MockerFixture):
    """A slash-command interaction whose responses and log-channel posts are recorded."""

    def __make(command="cmd", user=None):
        log_channel = mocker.Mock(spec=discord.TextChannel)
        log_channel.send = mocker.AsyncMock()
        original = mocker.Mock(edit=mocker.AsyncMock())

        interaction = mocker.Mock()
        interaction.command.name = command
        interaction.channel = mocker.Mock(spec=discord.TextChannel, mention="#mod-chat")
        interaction.guild.get_channel.return_value = log_channel
        interaction.user = user or mocker.Mock(
            spec=discord.Member,
            id=2,
            display_name="mod",
            display_avatar=mocker.Mock(url="https://a"),
        )
        interaction.created_at = datetime.datetime(2026, 1, 1, tzinfo=datetime.UTC)
        interaction.response.send_message = mocker.AsyncMock()
        interaction.response.defer = mocker.AsyncMock()
        interaction.response.edit_message = mocker.AsyncMock()
        interaction.response.send_modal = mocker.AsyncMock()
        interaction.original_response = mocker.AsyncMock(return_value=original)
        interaction.edit_original_response = mocker.AsyncMock()
        interaction.followup.send = mocker.AsyncMock()
        interaction.log_channel = log_channel
        interaction.original = original
        return interaction

    return __make
