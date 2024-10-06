import discord
from bot import ModdingwayBot
from settings import get_settings
import logging
from database import DatabaseConnection

settings = get_settings()

# TODO:
# create requirements.txt
# database directory for interacting with DB
# dockerize
# implementation logic
# error handling at the app level


if __name__ == "__main__":
    # Configure logging
    logging.basicConfig(level=settings.log_level)

    intents = discord.Intents.default()
    bot = ModdingwayBot(command_prefix="/", intents=intents)

    DatabaseConnection().connect()

    bot.run(settings.discord_token)
