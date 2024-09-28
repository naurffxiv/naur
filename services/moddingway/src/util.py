import discord

def log_info_and_embed(embed: discord.Embed, logger, message: str):
    embed.description += "\n" +  message
    logger.info(message)