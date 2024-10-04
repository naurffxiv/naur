import discord
from discord.ext.commands import Bot
from settings import get_settings
from commands.exile import create_exile_commands
import logging

settings = get_settings()
logger = logging.getLogger(__name__)


class ModdingwayBot(Bot):
    async def setup_hook(self):
        self._register_commands()

        guild = discord.Object(id=settings.guild_id)
        self.tree.copy_global_to(guild=guild)
        await self.tree.sync(guild=guild)

    async def on_ready(self):
        logger.info(f"Logged in as {self.user} (ID: {self.user.id})")

    def _register_commands(self):
        create_exile_commands(self)
