import logging

import discord
from discord.ext.commands import Bot

from moddingway.database import announcements_database
from moddingway.services import announcement_service
from moddingway.settings import get_settings
from moddingway.util import is_user_admin

from .helper import create_logging_embed, create_response_context

logger = logging.getLogger(__name__)

settings = get_settings()


class AnnouncementPublishView(discord.ui.View):
    def __init__(self, announcement_id, channel, interaction, bot, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.announcement_id = announcement_id
        self.channel = channel
        self.interaction = interaction
        self.bot = bot

    @discord.ui.button(label="Confirm", style=discord.ButtonStyle.success)
    async def confirm_button_callback(self, interaction, button):
        self.clear_items()
        await self.interaction.edit_original_response(view=self)
        async with create_response_context(interaction) as response_message:
            async with create_logging_embed(
                self.interaction,
                channel=self.channel,
                announcement_id=self.announcement_id,
            ) as logging_embed:
                await announcement_service.publish_announcement(
                    logging_embed,
                    channel=self.channel,
                    announcement_id=self.announcement_id,
                )
                response_message.set_string(
                    f"Successfully published announcement. (ID: {self.announcement_id})"
                )

    @discord.ui.button(label="Cancel", style=discord.ButtonStyle.danger)
    async def cancel_button_callback(self, interaction, button):
        self.clear_items()
        await self.interaction.edit_original_response(view=self)
        async with create_response_context(interaction) as response_message:
            response_message.set_string("Announcement publishing cancelled.")


def create_announcement_commands(bot: Bot) -> None:

    @bot.tree.command()
    @discord.app_commands.check(is_user_admin)
    @discord.app_commands.describe(announcement_text="Announcement goes here.")
    async def draft_announcement(
        interaction: discord.Interaction,
        announcement_text: discord.app_commands.Range[
            str, 1, 4000
        ],  # Client side char check
    ):
        """Draft an announcement"""
        async with create_response_context(interaction) as response_message:
            async with create_logging_embed(
                interaction, announcement_text
            ) as logging_embed:
                draft_id = await announcement_service.add_announcement(
                    logging_embed,
                    author=interaction.user,
                    announcement_text=announcement_text,
                    bot=bot,
                )

                response_message.set_string(
                    f"Successfully added announcement. (ID: {draft_id})"
                )

    @bot.tree.command()
    @discord.app_commands.check(is_user_admin)
    async def publish_announcement(
        interaction: discord.Interaction,
        channel: discord.TextChannel,
        announcement_id: int,
    ):
        """Publish an announcement"""
        announcement_json = announcements_database.get_announcement(
            announcement_id=announcement_id
        )
        if announcement_json is None:
            await interaction.response.send_message(
                "Announcement not found.", ephemeral=True
            )
        elif announcement_json["sent_flag"]:
            messageLink = (
                "https://discord.com/channels/"
                + str(settings.guild_id)
                + "/"
                + announcement_json["discord_msg_link"]
            )
            await interaction.response.send_message(
                f"Announcement already published: {messageLink}", ephemeral=True
            )
        else:
            await interaction.response.send_message(
                f"Are you sure you want to publish this announcement (ID: {announcement_id})",
                view=AnnouncementPublishView(
                    interaction=interaction,
                    channel=channel,
                    announcement_id=announcement_id,
                    bot=bot,
                ),
                ephemeral=True,
            )
