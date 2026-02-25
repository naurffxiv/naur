import logging
import re
from contextlib import asynccontextmanager
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta

import discord

from moddingway.constants import ERROR_MESSAGES, Role
from moddingway.settings import get_settings

settings = get_settings()
logger = logging.getLogger(__name__)


@dataclass
class EmbedField:
    name: str
    value: str


class UnableToNotify(RuntimeError):
    # NB: create a custom exception with RuntimeError as base
    # constructor/methods/attributes are all inherited
    pass


@asynccontextmanager
async def create_interaction_embed_context(
    log_channel: discord.abc.GuildChannel, **kwargs
):
    # optional args
    user = kwargs.get("user", None)
    description = kwargs.get("description", None)
    timestamp = kwargs.get("timestamp", None)
    footer = kwargs.get("footer", None)
    fields = kwargs.get("fields", None)

    embed = discord.Embed()
    try:
        if user is not None:
            embed.set_author(
                name=user.display_name,
                icon_url=user.display_avatar.url,
            )
        if description is not None:
            embed.description = description
        if timestamp is not None:
            embed.timestamp = timestamp
        if footer is not None:
            embed.set_footer(text=footer)
        if fields is not None:
            for field in fields:
                embed.add_field(name=field.name, value=field.value, inline=False)

        yield embed
    except Exception as e:
        embed.add_field(name="Error", value=f"{e!s}", inline=False)
        raise e
    finally:
        if isinstance(log_channel, discord.abc.Messageable):
            await log_channel.send(embed=embed)


def log_info_and_embed(embed: discord.Embed, logger, message: str):
    """
    Write a log message to both the default logger and add the string to
    the discord message that will be sent to the logging channel upon command
    finishing
    """
    if embed.description is None:
        embed.description = ""
    embed.description += "\n" + message
    logger.info(message)


def log_info_and_add_field(embed: discord.Embed, logger, name: str, value: str):
    """
    Write a log message to both the default logger and add the string to
    the discord message that will be sent to the logging channel upon command
    finishing
    """
    embed.add_field(name=name, value=value, inline=False)
    logger.info(f"{name}: {value}")


def _split_chunks(message_content: str, from_index: int, max_chunk_length: int = 2000):
    max_index = from_index + max_chunk_length

    # If remaining message is shorter than or equal to max chunk length, return the end
    if len(message_content) <= max_index:
        return len(message_content)

    # split based on newline
    newline_index = message_content.rfind("\n", from_index, max_index)
    if newline_index != -1:
        return newline_index + 1

    # split based on spaces
    space_index = message_content.rfind(" ", from_index, max_index)
    if space_index != -1:
        return space_index + 1

    # else just send a chunk of max_chunk_length characters
    # But wait, if message_content[max_index] is a space, we should consume it!
    if max_index < len(message_content) and message_content[max_index] in (" ", "\n"):
        return max_index + 1

    return max_index


def chunk_message(message_content: str, max_chunk_length: int = 2000):
    from_index = 0

    while from_index < len(message_content):
        to_index = _split_chunks(message_content, from_index, max_chunk_length)
        chunk = message_content[from_index:to_index].strip()
        if chunk:
            yield chunk
        from_index = to_index


async def send_chunked_message(channel: discord.abc.GuildChannel, message_content: str):
    for chunk in chunk_message(message_content):
        if isinstance(channel, discord.abc.Messageable):
            await channel.send(chunk)


async def send_dm(
    logging_embed: discord.Embed,
    member: discord.User | discord.Member,
    messageContent: str,
    context: str,
):
    try:
        channel = await member.create_dm()
        await channel.send(content=messageContent)
    except Exception as e:
        # Check if error has a code
        if hasattr(e, "code") and isinstance(e.code, int):
            code = e.code
            error_template_msg = ERROR_MESSAGES.get(
                code
            )  # Check if Error Code is present in constants.py
            if error_template_msg:
                formatted = error_template_msg.format(
                    user=getattr(member, "id", member), context=context
                )
                log_info_and_add_field(logging_embed, logger, "DM Status", formatted)
                return

        # Fallback to base error logging
        log_info_and_add_field(
            logging_embed,
            logger,
            "DM Status",
            f"Failed to send DM to <@{member.id}> for {context}, {e!s}",
        )


async def add_and_remove_role(
    member: discord.Member, role_to_add: Role, role_to_remove: Role
):
    discord_role_to_add = None
    discord_role_to_remove = None

    for role in member.guild.roles:
        if role.name == role_to_add.value:
            discord_role_to_add = role
        if role.name == role_to_remove.value:
            discord_role_to_remove = role

    if discord_role_to_add is None:
        # This role does not exist, likely a misconfiguration of the server
        raise Exception(f"Role {role_to_add.value} not found in server")

    if discord_role_to_remove is None:
        # This role does not exist, likely a misconfiguration of the server
        raise Exception(f"Role {role_to_remove.value} not found in server")

    await member.remove_roles(discord_role_to_remove)
    await member.add_roles(discord_role_to_add)


def user_has_role(user: discord.Member, role: Role) -> bool:
    return any(
        discord_role for discord_role in user.roles if discord_role.name == role.value
    )


def calculate_time_delta(delta_string: str | None) -> timedelta | None:
    if not delta_string:
        return None

    regex = r"^(\d\d?)(sec|min|min|hour|day)"  # Matches (digit, digit?)(option of [sec, min, hour, day])

    result = re.search(regex, delta_string)

    if result:
        duration = int(result.group(1))
        unit = result.group(2)

        delta = None

        if unit == "sec":
            delta = timedelta(seconds=duration)
        elif unit == "min":
            delta = timedelta(minutes=duration)
        elif unit == "hour":
            delta = timedelta(hours=duration)
        elif unit == "day":
            delta = timedelta(days=duration)

        return delta

    return None


async def is_user_moderator(interaction: discord.Interaction):
    if isinstance(interaction.user, discord.Member):
        return user_has_role(interaction.user, Role.MOD)
    return False


def timestamp_to_epoch(timestamp: datetime | None) -> int | None:
    if timestamp is None:
        return None
    # If already timezone-aware, convert to UTC; otherwise, assume UTC
    if timestamp.tzinfo is None:
        timestamp = timestamp.replace(tzinfo=UTC)
    else:
        timestamp = timestamp.astimezone(UTC)
    return round(timestamp.timestamp())


# TODO: MOD-166 add this check every time we are using the logging_channel_id
# Try to get the logging channel for event logging
def get_log_channel(self):
    """
    Get the logging channel and handle errors if it doesn't exist.
    Returns the channel or None if not found.
    """
    log_channel = self.get_channel(settings.logging_channel_id)

    if log_channel is None:
        logger = logging.getLogger(__name__)
        logger.error(
            f"Logging channel {settings.logging_channel_id} not found. Event will not be logged to Discord."
        )

    return log_channel


# New utility functions for member events


async def find_and_assign_role(
    member: discord.Member, role_enum: Role
) -> tuple[bool, str, discord.Role | None]:
    """
    Finds a role by enum and assigns it to a member.

    Args:
        member: The member to assign the role to
        role_enum: The role enum to find and assign

    Returns:
        tuple containing:
            - is_role_assigned: Whether the role was successfully assigned
            - message: A descriptive message about the result
            - role: The role object if found, None otherwise
    """
    try:
        # Find the role by name using the enum value
        role = discord.utils.get(member.guild.roles, name=role_enum.value)

        if role is None:
            error_msg = f"Role '{role_enum.value}' not found in the server."
            logger.error(error_msg)
            return False, error_msg, None

        # Assign the role to the member
        await member.add_roles(role)
        success_msg = f"Successfully assigned <@&{role.id}> role"
        logger.debug(success_msg)
        return True, success_msg, role
    except discord.Forbidden:
        error_msg = (
            f"Bot does not have permission to add roles to {member.display_name}"
        )
        logger.error(error_msg)
        return False, error_msg, None
    except Exception as e:
        error_msg = f"Failed to assign role: {e!s}"
        logger.error(error_msg)
        return False, error_msg, None

    # Adding to test script
