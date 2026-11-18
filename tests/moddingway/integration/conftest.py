"""
Integration tests run against a real Postgres (the `postgres_test` compose
service). Discord is replaced by the small in-memory fakes below; the database is
never mocked.
"""

from collections.abc import Callable
from datetime import UTC, datetime
from types import SimpleNamespace
from typing import Any

import discord
import pytest

from moddingway.constants import Role
from moddingway.database import DatabaseConnection
from moddingway.services import exiles

VERIFIED_ROLE_ID = 1
EXILED_ROLE_ID = 2
NON_VERIFIED_ROLE_ID = 3
MOD_ROLE_ID = 4
STICKY_ROLE_IDS = [101, 102]

TABLES = ["users", "exiles", "strikes", "notes", "roles", "forms"]


class FakeDiscordError(Exception):
    pass


class FakeRole:
    def __init__(self, role_id: int, name: str):
        self.id = role_id
        self.name = name

    def __repr__(self):
        return f"FakeRole({self.name})"

    def is_default(self) -> bool:
        return False


class FakeLogChannel(discord.abc.Messageable):
    """The logging channel: collects the embeds the bot sends to it."""

    def __init__(self):
        self.embeds: list[discord.Embed] = []
        self.fails = False

    async def send(self, *, embed: discord.Embed):  # type: ignore[override]
        if self.fails:
            raise FakeDiscordError("Missing access to the logging channel")
        self.embeds.append(embed)


class FakeGuild:
    def __init__(self, roles: list[FakeRole]):
        self.roles = roles
        self.members: dict[int, FakeMember] = {}
        self.log_channel: FakeLogChannel | None = None

    def get_role(self, role_id: int) -> FakeRole | None:
        return next((role for role in self.roles if role.id == role_id), None)

    def role(self, name: Role) -> FakeRole:
        return next(role for role in self.roles if role.name == name.value)

    def get_member(self, member_id: int) -> FakeMember | None:
        return self.members.get(member_id)

    def get_channel(self, channel_id: int) -> FakeLogChannel | None:
        # Without a logging channel, event handlers skip sending embeds
        return self.log_channel


class FakeDMChannel:
    def __init__(self):
        self.sent: list[str] = []

    async def send(self, content: str):
        self.sent.append(content)


class FakeMember:
    """
    Stand-in for discord.Member. ``failing_role_ids`` makes any change to those
    roles (add, remove or edit) raise, like a Discord permission error would.
    """

    def __init__(self, member_id: int, guild: FakeGuild, roles: list[FakeRole]):
        self.id = member_id
        self.guild = guild
        self.roles = list(roles)
        self.display_name = f"member-{member_id}"
        self.mention = f"<@{member_id}>"
        self.display_avatar = SimpleNamespace(url="https://example.invalid/a.png")
        self.created_at = datetime(2020, 1, 1, tzinfo=UTC)
        self.failing_role_ids: set[int] = set()
        self.dms_open = True
        self.dm_channel = FakeDMChannel()

    @property
    def __class__(self):  # type: ignore[override]
        # Pass isinstance(member, discord.Member) checks, as Mock(spec=...) does
        return discord.Member

    def role_names(self) -> set[str]:
        return {role.name for role in self.roles}

    def role_ids(self) -> set[int]:
        return {role.id for role in self.roles}

    def _check(self, roles: tuple[FakeRole, ...]):
        failing = [role for role in roles if role.id in self.failing_role_ids]
        if failing:
            raise FakeDiscordError(f"Missing permissions for {failing}")

    async def add_roles(self, *roles: FakeRole):
        self._check(roles)
        for role in roles:
            if role not in self.roles:
                self.roles.append(role)

    async def remove_roles(self, *roles: FakeRole):
        self._check(roles)
        self.roles = [role for role in self.roles if role not in roles]

    async def edit(self, *, roles: list[FakeRole]):
        changed = set(roles).symmetric_difference(self.roles)
        self._check(tuple(changed))
        self.roles = list(roles)

    async def create_dm(self) -> FakeDMChannel:
        if not self.dms_open:
            raise FakeDiscordError("Cannot send messages to this user")
        return self.dm_channel


class FakeOriginalResponse:
    def __init__(self):
        self.content: str | None = None

    async def edit(self, *, content: str):
        self.content = content


class FakeInteractionResponse:
    def __init__(self):
        # (content, keyword arguments) of each send_message call
        self.sent: list[tuple[str, dict[str, Any]]] = []

    async def send_message(self, content: str, **kwargs):
        self.sent.append((content, kwargs))


class FakeInteraction:
    """
    Stand-in for discord.Interaction. ``reply`` is the final text of the
    "Processing..." message the bot edits, ``response.sent`` every direct reply.
    """

    def __init__(self, user, guild: FakeGuild, command_name: str | None = None):
        self.user = user
        self.guild = guild
        self.guild_id = 1
        self.command = (
            SimpleNamespace(name=command_name) if command_name is not None else None
        )
        self.channel = SimpleNamespace(mention="#general")
        self.created_at = datetime.now(UTC)
        self.response = FakeInteractionResponse()
        self.original = FakeOriginalResponse()
        self.edited_views: list[Any] = []

    async def original_response(self) -> FakeOriginalResponse:
        return self.original

    async def edit_original_response(self, *, view):
        self.edited_views.append(view)

    @property
    def reply(self) -> str | None:
        return self.original.content


class FakeCommandTree:
    def __init__(self):
        self.commands: dict[str, Callable[..., Any]] = {}

    def command(self):
        def __register(callback):
            self.commands[callback.__name__] = callback
            return callback

        return __register


class FakeBot:
    """Collects the commands and events the bot registers, by name."""

    def __init__(self, guild: FakeGuild):
        self.guild = guild
        self.tree = FakeCommandTree()
        self.events: dict[str, Callable[..., Any]] = {}

    def event(self, handler):
        self.events[handler.__name__] = handler
        return handler

    def get_guild(self, guild_id: int) -> FakeGuild | None:
        return self.guild

    def get_channel(self, channel_id: int) -> FakeLogChannel | None:
        return self.guild.get_channel(channel_id)


@pytest.fixture(scope="session")
def database() -> DatabaseConnection:
    connection = DatabaseConnection()
    connection.create_tables()
    return connection


def _truncate(database: DatabaseConnection):
    with database.get_cursor() as cursor:
        cursor.execute(f"TRUNCATE {', '.join(TABLES)} RESTART IDENTITY CASCADE")


@pytest.fixture(autouse=True)
def clean_database(database: DatabaseConnection):
    _truncate(database)
    yield
    _truncate(database)


@pytest.fixture
def query(database: DatabaseConnection) -> Callable[..., list[tuple]]:
    def __query(sql: str, *params) -> list[tuple]:
        with database.get_cursor() as cursor:
            cursor.execute(sql, params)
            return cursor.fetchall()

    return __query


@pytest.fixture(autouse=True)
def sticky_roles(monkeypatch) -> list[int]:
    monkeypatch.setattr(exiles.settings, "sticky_roles", STICKY_ROLE_IDS)
    return STICKY_ROLE_IDS


@pytest.fixture
def guild() -> Any:
    return FakeGuild(
        [
            FakeRole(VERIFIED_ROLE_ID, Role.VERIFIED.value),
            FakeRole(EXILED_ROLE_ID, Role.EXILED.value),
            FakeRole(NON_VERIFIED_ROLE_ID, Role.NON_VERIFIED.value),
            FakeRole(MOD_ROLE_ID, Role.MOD.value),
            *(FakeRole(role_id, f"sticky-{role_id}") for role_id in STICKY_ROLE_IDS),
        ]
    )


@pytest.fixture
def make_member(guild: FakeGuild) -> Callable[..., Any]:
    """Create a member in the guild. Returned as Any so it can stand in for
    discord.Member in calls to the bot's code."""

    def __make_member(
        member_id: int = 12345,
        roles: list[Role] | None = None,
        role_ids: list[int] | None = None,
    ) -> Any:
        if roles is None:
            roles = [Role.VERIFIED]
        member_roles = [guild.role(role) for role in roles]
        member_roles += [
            role for role_id in role_ids or [] if (role := guild.get_role(role_id))
        ]
        member = FakeMember(member_id, guild, member_roles)
        guild.members[member_id] = member
        return member

    return __make_member


@pytest.fixture
def log_channel(guild: FakeGuild) -> FakeLogChannel:
    guild.log_channel = FakeLogChannel()
    return guild.log_channel


@pytest.fixture
def bot(guild: FakeGuild) -> Any:
    """A bot for the guild. Returned as Any so it can stand in for the real Bot."""
    return FakeBot(guild)
