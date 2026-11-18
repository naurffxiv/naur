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


def _dm_failure_message(
    member: discord.User | discord.Member, context: str, error: Exception
) -> str:
    """Describe why a DM could not be delivered, for the logging channel."""
    # Check if error has a code
    code = getattr(error, "code", None)
    if isinstance(code, int):
        # Check if Error Code is present in constants.py
        error_template_msg = ERROR_MESSAGES.get(code)
        if error_template_msg:
            return error_template_msg.format(
                user=getattr(member, "id", member), context=context
            )

    # Fallback to base error message
    return f"Failed to send DM to <@{member.id}> for {context}, {error!s}"


async def try_send_dm(
    member: discord.User | discord.Member, message_content: str, context: str
) -> str | None:
    """Send a DM, returning a failure description instead of raising."""
    try:
        channel = await member.create_dm()
        await channel.send(content=message_content)
    except Exception as e:
        return _dm_failure_message(member, context, e)
    return None


async def send_dm(
    logging_embed: discord.Embed,
    member: discord.User | discord.Member,
    messageContent: str,
    context: str,
):
    failure = await try_send_dm(member, messageContent, context)
    if failure is not None:
        log_info_and_add_field(logging_embed, logger, "DM Status", failure)


def get_guild_role(guild: discord.Guild, role: Role) -> discord.Role | None:
    return discord.utils.get(guild.roles, name=role.value)


async def add_and_remove_role(
    member: discord.Member, role_to_add: Role, role_to_remove: Role
):
    """
    Swap one role for another in a single Discord call, so the change is either
    fully applied or not at all. Every other role the member has is kept.
    """
    discord_role_to_add = get_guild_role(member.guild, role_to_add)
    if discord_role_to_add is None:
        # This role does not exist, likely a misconfiguration of the server
        raise Exception(f"Role {role_to_add.value} not found in server")

    discord_role_to_remove = get_guild_role(member.guild, role_to_remove)
    if discord_role_to_remove is None:
        # This role does not exist, likely a misconfiguration of the server
        raise Exception(f"Role {role_to_remove.value} not found in server")

    # @everyone is implicit and cannot be part of the new role list
    new_roles = [
        role
        for role in member.roles
        if role != discord_role_to_remove and not role.is_default()
    ]
    if discord_role_to_add not in new_roles:
        new_roles.append(discord_role_to_add)
    await member.edit(roles=new_roles)


def user_has_role(user: discord.Member, role: Role) -> bool:
    return any(
        discord_role for discord_role in user.roles if discord_role.name == role.value
    )


# Matches (digit, digit?)(option of [sec, min, hour, day])
DURATION_REGEX = re.compile(r"^(\d\d?)(sec|min|hour|day)")
_DURATION_UNITS = {"sec": "second", "min": "minute", "hour": "hour", "day": "day"}


def parse_duration(duration_string: str | None) -> tuple[int, str] | None:
    """Split a duration such as "12hour" into its amount and unit (12, "hour").

    The unit is returned as a full word: second, minute, hour or day.
    """
    if not duration_string:
        return None
    result = DURATION_REGEX.search(duration_string)
    if result is None:
        return None
    return int(result.group(1)), _DURATION_UNITS[result.group(2)]


def calculate_time_delta(delta_string: str | None) -> timedelta | None:
    parsed = parse_duration(delta_string)
    if parsed is None:
        return None

    amount, unit = parsed
    return timedelta(**{f"{unit}s": amount})


async def is_user_moderator(interaction: discord.Interaction):
    if not isinstance(interaction.user, discord.Member):
        return False
    if user_has_role(interaction.user, Role.MOD) or user_has_role(
        interaction.user, Role.ADMIN
    ):
        return True
    raise discord.app_commands.MissingAnyRole([Role.MOD, Role.ADMIN])


async def is_user_admin(interaction: discord.Interaction):
    if not isinstance(interaction.user, discord.Member):
        return False
    if user_has_role(interaction.user, Role.ADMIN):
        return True
    raise discord.app_commands.MissingRole(Role.ADMIN)


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
        role = get_guild_role(member.guild, role_enum)

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
