import discord
from collections.abc import Coroutine
from settings import get_settings

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
