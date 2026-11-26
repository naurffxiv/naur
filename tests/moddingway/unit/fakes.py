"""Small builders shared by the unit tests (importable because conftest.py lives here)."""

import discord

from moddingway.constants import UserRole
from moddingway.database.models import User


def embed_fields(embed: discord.Embed) -> list[tuple[str, str]]:
    return [(field.name, field.value) for field in embed.fields]


def make_user(
    user_id=1, role=UserRole.USER, temporary=0, permanent=0, banned=False
) -> User:
    return User(
        user_id=user_id,
        discord_user_id=str(user_id),
        discord_guild_id="1",
        user_role=role,
        temporary_points=temporary,
        permanent_points=permanent,
        is_banned=banned,
    )
