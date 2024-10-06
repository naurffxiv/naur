import discord
from settings import get_settings
from contextlib import asynccontextmanager

settings = get_settings()


@asynccontextmanager
async def create_logging_embed(interaction: discord.Interaction):
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

        yield embed

    except Exception as e:
        embed.add_field(name="Error", value=e)
        raise e
    finally:
        await log_channel.send(embed=embed)
