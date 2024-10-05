import discord
from collections.abc import Coroutine
from settings import get_settings
import enums

settings = get_settings()


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


async def send_dm(member: discord.Member, messageContent: str):
    channel = await member.create_dm()
    await channel.send(content=messageContent)


# NB this feels very messy, fragile, and easy to mess up
# There might be a better way to wrap all this in something cleaner
async def run_command_with_logging(interaction, command: Coroutine, *args):
    """
    Wrap a call to business logic with an automatic creation of a message
    inside the server's logging channel
    """
    embed = discord.Embed()
    log_channel = interaction.guild.get_channel(settings.logging_channel_id)
    try:
        embed.set_author(
            name=interaction.user.display_name,
            icon_url=interaction.user.display_avatar.url,
        )
        embed.timestamp = interaction.created_at
        embed.description = f"Used `{interaction.command.name}` command in {interaction.channel.mention}"
        embed.add_field(name="Action", value=f"/{interaction.command.name}")

        await command(embed, *args)

    except Exception as e:
        embed.add_field(name="Error", value=e)
    finally:
        await log_channel.send(embed=embed)


async def add_and_remove_role(
    member: discord.Member, role_to_add: enums.Role, role_to_remove: enums.Role
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


def user_has_role(user: discord.Member, role: enums.Role) -> bool:
    return any(
        discord_role
        for discord_role in user.guild.roles
        if discord_role.name == role.value
    )
