import logging
import math

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


class AnnouncementPaginator(
    discord.ui.View
):  ##TODO: grey out buttons if not functional ?
    def __init__(self, data, author):
        super().__init__(timeout=60)  # Buttons disable after 60s of inactivity
        self.data = data
        self.author = author
        self.current_page = 1
        self.per_page = 8
        self.total_pages = math.ceil(len(data) / self.per_page)
        self.message: discord.Message | None = None

    def create_embed(self):
        start = (self.current_page - 1) * self.per_page
        end = start + self.per_page
        page_data = self.data[start:end]

        embed = discord.Embed(title="Announcements")

        for row in page_data:
            announcement_id, revisions, sent_flag, discord_message_link = row
            messageLink = (
                "https://discord.com/channels/"
                + str(settings.guild_id)
                + "/"
                + str(discord_message_link)
            )
            status = f"Sent: {messageLink}" if sent_flag else "Unsent"

            embed.add_field(
                name=f"Announcement ID: {announcement_id}",
                value=revisions[-1][
                    "content"
                ],  ##TODO: this needs to be trimmed too many long announcements would hit the embed limit
                inline=True,
            )
            embed.add_field(
                name="Latest Revision",
                value=f"<@{revisions[-1]['author_id']}>",
                inline=True,
            )
            embed.add_field(
                name="Status",
                value=status,
                inline=True,
            )

        embed.set_footer(text=f"Page {self.current_page}/{self.total_pages}")
        return embed
    ##TODO: buttons go unresponsivbe for a while when you click on them on the end pages, make it more responsive
    @discord.ui.button(label="Previous", style=discord.ButtonStyle.gray)
    async def prev_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        if self.current_page > 1:
            self.current_page -= 1
            await interaction.response.edit_message(
                embed=self.create_embed(), view=self
            )

    @discord.ui.button(label="Next", style=discord.ButtonStyle.gray)
    async def next_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        if self.current_page < self.total_pages:
            self.current_page += 1
            await interaction.response.edit_message(
                embed=self.create_embed(), view=self
            )


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

    @bot.tree.command()
    @discord.app_commands.check(is_user_admin)
    async def list_announcements(
        interaction: discord.Interaction,
        ephemeral: bool,
        sent_status: bool | None = None,
    ):
        """List announcements"""
        announcement_list = await announcement_service.list_announcements_service(
            status=sent_status
        )

        view = AnnouncementPaginator(announcement_list, interaction.user)
        await interaction.response.send_message(
            embed=view.create_embed(), view=view, ephemeral=ephemeral
        )
        view.message = await interaction.original_response()

    @bot.tree.command()
    @discord.app_commands.check(is_user_admin)
    async def show_announcement(
        interaction: discord.Interaction,
        announcement_id: int,
    ):
        """Show announcement"""

        announcement_json = announcements_database.get_announcement(
            announcement_id=announcement_id
        )
        if announcement_json is None:
            await interaction.response.send_message(
                "Announcement not found.", ephemeral=True
            )
        else:
            messageLink = (
                "https://discord.com/channels/"
                + str(settings.guild_id)
                + "/"
                + str(announcement_json["discord_msg_link"])
            )
            status = (
                f"Sent: {messageLink}" if announcement_json["sent_flag"] else "Unsent"
            )
            announcement_embed = discord.Embed(
                title="Announcement Draft",
                description=announcement_json["revisions"][-1]["content"],
            )
            announcement_embed.add_field(
                name="Latest Revision",
                value=f"<@{announcement_json['revisions'][-1]['author_id']}>",
                inline=True,
            )
            announcement_embed.add_field(name="Status", value=status, inline=True)

            announcement_embed.set_footer(
                text=f"Announcement ID {announcement_json['announcement_id']}"
            )

            await interaction.response.send_message(embed=announcement_embed)
