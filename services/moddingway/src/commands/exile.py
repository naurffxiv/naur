import discord
from discord.ext.commands import Bot
from settings import get_settings
from services.exile_service import exile_user, unexile_user
from collections.abc import Coroutine

settings = get_settings()

def create_unexile_command(bot: Bot) -> None:
    @bot.tree.command()
    @discord.app_commands.describe(
        user="User being exiled"
    )
    async def unexile(interaction: discord.Interaction, user: discord.Member):
        """Exile the specified user."""

        await run_command_with_logging(interaction, unexile_user, user)

        await interaction.response.send_message(f"Successfully unexiled {user.mention}", ephemeral=True)

def create_exile_command(bot: Bot) -> None:
    @bot.tree.command()
    @discord.app_commands.describe(
        user="User being exiled",
        duration="The duration of the exile. TBA format",
        reason="Reason for exile"
    )
    async def exile(interaction: discord.Interaction, user: discord.Member, duration: str, reason: str):
        """Exile the specified user."""

        await run_command_with_logging(interaction, exile_user, user, duration, reason)

        await interaction.response.send_message(f"Successfully exiled {user.mention}", ephemeral=True)


# NB this feels very messy, fragile, and easy to mess up
# There might be a better way to wrap all this in something cleaner and better reusable
async def run_command_with_logging(interaction, command: Coroutine, *args):
    embed = discord.Embed()
    try:
        embed.set_author(name=interaction.user.display_name, icon_url=interaction.user.display_avatar.url)
        embed.timestamp = interaction.created_at
        embed.description = f"Used `{interaction.command.name}` command in {interaction.channel.mention}"
        embed.add_field(name="Action", value=f"/{interaction.command.name}")

        await command(embed, *args)

        log_channel = interaction.guild.get_channel(settings.logging_channel_id)
    except Exception as e:
        embed.add_field(name="Error", value=e)
    finally:
        await log_channel.send(embed=embed)

