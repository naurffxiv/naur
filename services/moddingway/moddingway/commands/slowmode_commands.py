import logging

import discord
from discord.ext.commands import Bot

from moddingway.services.slowmode_service import edit_slowmode
from moddingway.settings import get_settings
from moddingway.util import is_user_moderator

from .helper import create_logging_embed, create_response_context

settings = get_settings()
logger = logging.getLogger(__name__)


def create_slowmode_commands(bot: Bot) -> None:
    @bot.tree.command()
    @discord.app_commands.check(is_user_moderator)
    @discord.app_commands.describe(
        interval="The number of seconds desired between each message. Must be between 0 (off) and 21600 (6 hours).",
        channel="Channel name slowmode is being modified in",
    )
    async def set_slowmode(
        interaction: discord.Interaction,
        interval: int,
        channel: discord.TextChannel | None = None,
    ):
        """Set the slowmode interval for the specified channel."""
        target_channel: discord.TextChannel
        if channel is None:  # set default channel
            if isinstance(interaction.channel, discord.TextChannel):
                target_channel = interaction.channel
            else:
                await interaction.response.send_message(
                    "This command can only be used in a text channel.", ephemeral=True
                )
                return
        else:
            target_channel = channel

        async with create_response_context(interaction) as response_message:
            async with create_logging_embed(
                interaction, interval=interval, channel=channel
            ) as logging_embed:
                result = await edit_slowmode(
                    logging_embed=logging_embed,
                    channel=target_channel,
                    interval=interval,
                )
                response_message.set_string(result)
