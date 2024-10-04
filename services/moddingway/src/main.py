import discord
from bot import ModdingwayBot
from settings import get_settings
import logging

settings = get_settings()

# TODO:
# update readme for python
# create requirements.txt
# python linting of some kind
# Permissions on commands
# database directory for interacting with DB
# dockerize
# implementation logic


if __name__ == "__main__":
    # Configure logging
    logging.basicConfig(level=settings.log_level)

    intents = discord.Intents.default()
    bot = ModdingwayBot(command_prefix="/", intents=intents)

    bot.run(settings.discord_token)
