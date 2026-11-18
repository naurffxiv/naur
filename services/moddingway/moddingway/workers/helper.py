import discord

from moddingway.util import (
    create_interaction_embed_context,
    get_log_channel,
)


def create_autounexile_embed(
    self,
    user: discord.Member | None,
    discord_id: int,
    exile_id: str,
    end_timestamp: str,
):
    return create_interaction_embed_context(
        get_log_channel(self),
        user=user,
        timestamp=end_timestamp,
        description=f"<@{discord_id}>'s exile has timed out",
        footer=f"Exile ID: {exile_id}",
    )
