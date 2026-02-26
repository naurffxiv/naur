import discord
from discord.ext.commands import Bot

from moddingway.services.send_message_service import send_message_to_channel
from moddingway.util import is_user_moderator

from .helper import create_logging_embed, create_response_context


def create_send_message_commands(bot: Bot) -> None:
    @bot.tree.command(name="send_message")
    @discord.app_commands.check(is_user_moderator)
    @discord.app_commands.describe(
        channel="Channel to post message in",
        message="Message to post (max. 250 characters)",
    )
    async def send_message(
        interaction: discord.Interaction,
        channel: discord.TextChannel,
        message: str,
    ):
        async with create_response_context(interaction) as response_message:
            async with create_logging_embed(
                interaction, channel=channel, message=message
            ) as logging_embed:
                result = await send_message_to_channel(logging_embed, channel, message)
                response_message.set_string(result)
