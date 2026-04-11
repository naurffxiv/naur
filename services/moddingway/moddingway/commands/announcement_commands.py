import logging
import math
import math

import discord
from discord.ext.commands import Bot

from moddingway.constants import MAX_CHAR_LIMIT, MESSAGE_LINK_PREFIX, TRUNCATE_OFFSET
from moddingway.constants import MAX_CHAR_LIMIT, MESSAGE_LINK_PREFIX, TRUNCATE_OFFSET
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


class AnnouncementPaginator(discord.ui.View):
    def __init__(self, data, sent_status, author):
        super().__init__(timeout=60)  # Buttons disable after 60s of inactivity
        self.data = data
        self.author = author
        self.sent_status = sent_status
        self.current_page = 1
        self.per_page = 6
        self.total_pages = math.ceil(len(data) / self.per_page)
        self.message: discord.Message | None = None
        self.update_button_states()

    def create_embed(self):
        start = (self.current_page - 1) * self.per_page
        end = start + self.per_page
        page_data = self.data[start:end]
        sent_status = self.sent_status

        if sent_status is True:
            status_title = "Sent"
        elif sent_status is False:
            status_title = "Unsent"
        else:
            status_title = "All"
        embed = discord.Embed(title=f"{status_title} Announcements")

        for row in page_data:
            announcement_id, revisions, sent_flag, discord_message_link = row
            shortened_rev = ""
            if (
                len(revisions[-1]["content"]) > MAX_CHAR_LIMIT
            ):  # check if revision is over 800 chars, trim it down and add "..." to end to show that its cut off
                shortened_rev = (
                    revisions[-1]["content"][: MAX_CHAR_LIMIT - TRUNCATE_OFFSET] + "..."
                )
            messageLink = (
                MESSAGE_LINK_PREFIX
                + str(settings.guild_id)
                + "/"
                + str(discord_message_link)
            )
            status = f"Sent: {messageLink}" if sent_flag else "Unsent"

            embed.add_field(
                name="ID:",
                value=announcement_id,
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
            embed.add_field(
                name=" ",
                value=shortened_rev,  # 1024 char limit here per value and 6000 for total chars in whole embed so limiting to 800 per announcement preview should be fine
                inline=False,
            )

        embed.set_footer(text=f"Page {self.current_page}/{self.total_pages}")
        return embed

    def update_button_states(self):
        # Access the decorated buttons directly
        self.prev_button.disabled = self.current_page <= 1
        self.next_button.disabled = self.current_page >= self.total_pages

    @discord.ui.button(label="Previous", style=discord.ButtonStyle.gray, disabled=True)
    async def prev_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        if self.current_page <= 1:
            return await interaction.response.defer()

        self.current_page -= 1
        self.update_button_states()
        await interaction.response.edit_message(embed=self.create_embed(), view=self)

    @discord.ui.button(label="Next", style=discord.ButtonStyle.gray)
    async def next_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        if self.current_page >= self.total_pages:
            return await interaction.response.defer()

        self.current_page += 1
        self.update_button_states()
        await interaction.response.edit_message(embed=self.create_embed(), view=self)


class AnnouncementEditModal(discord.ui.Modal, title="Edit Announcement"):
    announcement_content = discord.ui.TextInput(
        label="Announcement Text",
        style=discord.TextStyle.long,
        placeholder="Edit your announcement here...",
        max_length=4000,
    )

    def __init__(
        self, announcement_id: int, current_content: str, bot: Bot, *args, **kwargs
    ):
        super().__init__(*args, **kwargs)
        self.announcement_id = announcement_id
        self.bot = bot
        self.announcement_content.default = current_content

    async def on_submit(self, interaction: discord.Interaction):
        new_content = self.announcement_content.value

        announcements_database.add_revision(
            announcement_id=self.announcement_id,
            author_id=interaction.user.id,
            content=new_content,
        )

        await interaction.response.send_message(
            f"Successfully updated announcement (ID: {self.announcement_id}).",
            ephemeral=True,
        )


class AnnouncementShowView(discord.ui.View):
    def __init__(self, announcement_json, author, bot):
        super().__init__(timeout=180)
        self.announcement_json = announcement_json
        self.revisions = announcement_json.get("revisions", [])
        self.current_rev_index = len(self.revisions) - 1
        self.author = author
        self.bot = bot
        self.update_button_states()

    def create_embed(self):
        rev = self.revisions[self.current_rev_index]
        messageLink = (
            MESSAGE_LINK_PREFIX
            + str(settings.guild_id)
            + "/"
            + str(self.announcement_json["discord_msg_link"])
        )
        status = (
            f"Sent: {messageLink}" if self.announcement_json["sent_flag"] else "Unsent"
        )

        embed = discord.Embed(
            title=f"Announcement Draft (Rev {self.current_rev_index + 1}/{len(self.revisions)})",
            description=rev["content"],
        )
        embed.add_field(
            name="Revision Author",
            value=f"<@{rev['author_id']}>",
            inline=True,
        )
        embed.add_field(name="Status", value=status, inline=True)
        embed.set_footer(
            text=f"Announcement ID {self.announcement_json['announcement_id']}"
        )
        return embed

    def update_button_states(self):
        self.left_button.disabled = self.current_rev_index <= 0
        self.right_button.disabled = self.current_rev_index >= len(self.revisions) - 1

    @discord.ui.button(label="◀", style=discord.ButtonStyle.gray)
    async def left_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        if self.current_rev_index > 0:
            self.current_rev_index -= 1
            self.update_button_states()
            await interaction.response.edit_message(
                embed=self.create_embed(), view=self
            )

    @discord.ui.button(label="Edit", style=discord.ButtonStyle.primary)
    async def edit_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        current_content = self.revisions[self.current_rev_index]["content"]
        modal = AnnouncementEditModal(
            announcement_id=self.announcement_json["announcement_id"],
            current_content=current_content,
            bot=self.bot,
        )
        await interaction.response.send_modal(modal)

    @discord.ui.button(label="▶", style=discord.ButtonStyle.gray)
    async def right_button(
        self, interaction: discord.Interaction, button: discord.ui.Button
    ):
        if interaction.user != self.author:
            return await interaction.response.send_message(
                "This isn't your menu!", ephemeral=True
            )

        if self.current_rev_index < len(self.revisions) - 1:
            self.current_rev_index += 1
            self.update_button_states()
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
        """Draft an announcement [4000 character limit]"""
        """Draft an announcement [4000 character limit]"""
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
        if len(announcement_list) == 0:
            await interaction.response.send_message(
                "Can't find announcements with selected parameters.", ephemeral=True
            )
        else:
            view = AnnouncementPaginator(
                announcement_list, sent_status, interaction.user
            )
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
            view = AnnouncementShowView(
                announcement_json=announcement_json, author=interaction.user, bot=bot
            )
            await interaction.response.send_message(
                embed=view.create_embed(), view=view
            )
