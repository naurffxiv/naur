import discord
from bot import ModdingwayBot
from settings import get_settings
import logging

settings = get_settings()

# TODO:
# create requirements.txt
# Permissions on commands
# database directory for interacting with DB
# dockerize
# implementation logic
# error handling at the app level


if __name__ == "__main__":
    # Configure logging
    logging.basicConfig(level=settings.log_level)

    intents = discord.Intents.default()
    bot = ModdingwayBot(command_prefix="/", intents=intents)

    bot.run(settings.discord_token)
