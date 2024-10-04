import discord
from discord.ext.commands import Bot, CommandError
from settings import get_settings
from services.exile_service import exile_user, unexile_user
from util import run_command_with_logging


settings = get_settings()

async def is_specific_person(interaction: discord.Interaction):
    if interaction.application_id != 1234567890:
        raise CommandError("You do not have permission to run this command")

def create_exile_commands(bot: Bot) -> None:
    @bot.tree.command()
    @discord.app_commands.check(is_specific_person)
    @discord.app_commands.describe(user="User being exiled")
    async def unexile(interaction: discord.Interaction, user: discord.Member):
        """Unexile the specified user."""

        await run_command_with_logging(interaction, unexile_user, user)

        await interaction.response.send_message(
            f"Successfully unexiled {user.mention}", ephemeral=True
        )

    @bot.tree.command()
    @discord.app_commands.describe(
        user="User being exiled",
        duration="The duration of the exile. TBA format",
        reason="Reason for exile",
    )
    async def exile(
        interaction: discord.Interaction,
        user: discord.Member,
        duration: str,
        reason: str,
    ):
        """Exile the specified user."""

        await run_command_with_logging(interaction, exile_user, user, duration, reason)

        await interaction.response.send_message(
            f"Successfully exiled {user.mention}", ephemeral=True
        )
