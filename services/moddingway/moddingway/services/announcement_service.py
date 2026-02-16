import logging

import discord

from moddingway.database import announcements_database
from moddingway.database.models import AnnouncementRevision
from moddingway.settings import get_settings
from moddingway.util import log_info_and_add_field

logger = logging.getLogger(__name__)
settings = get_settings()


async def add_announcement(
    logging_embed: discord.Embed,
    author: discord.User | discord.Member,
    announcement_text: str,
    bot: discord.Client,
) -> int:

    announcement_rev = AnnouncementRevision(
        version=0, content=announcement_text, author_id=author.id
    )

    new_id = announcements_database.add_announcement(announcement_rev=announcement_rev)
    logging_embed.set_footer(text=f"Announcement ID: {new_id}")

    test_channel_id = settings.announcement_draft_channel
    channel = bot.get_channel(test_channel_id)

    if channel is None:
        raise ValueError(
            f"Could not find announcement channel with ID {test_channel_id}"
        )

    announcement_embed = discord.Embed(
        title="Announcement Draft",
        description=announcement_text,
    )
    announcement_embed.add_field(name="Edited By", value=f"<@{author.id}>", inline=True)
    announcement_embed.add_field(name="Status", value="Unsent", inline=True)

    announcement_embed.set_footer(text=f"Announcement ID {new_id}")
    sent_message = await channel.send(embed=announcement_embed)
    if sent_message:
        log_info_and_add_field(
            logging_embed,
            logger,
            "Result",
            f"Announcement draft created, ID:{new_id}",
        )

    return new_id


## should sent announcement be able to get sent again?


async def publish_announcement(logging_embed, channel, announcement_id, bot) -> bool:

    announcement_json = announcements_database.get_announcement(
        announcement_id=announcement_id
    )

    if announcement_json is None:
        return False
    revisions = announcement_json.get("revisions", [])
    latest_revision_dict = revisions[-1]
    latest_revision = AnnouncementRevision(**latest_revision_dict)
    publish_embed = discord.Embed(
        description=latest_revision.content
    )  # put it in embed so it can support 4k characters
    sent_message = await channel.send(embed=publish_embed)
    if sent_message:
        # double check if it actually sets
        announcements_database.set_sent(
            announcement_id=announcement_id, discord_msg_id=sent_message.id
        )

        log_info_and_add_field(
            logging_embed, logger, "Result", "Published announcement"
        )
