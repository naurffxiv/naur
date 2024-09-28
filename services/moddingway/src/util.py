import discord

def log_info_and_embed(embed: discord.Embed, logger, message: str):
    """Write a log message to both the default logger and add the string to the discord message that will be sent to the logging channel upon command finishing"""
    embed.description += "\n" +  message
    logger.info(message)