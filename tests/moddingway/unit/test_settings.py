import logging

import pytest

from moddingway import settings
from moddingway.constants import AUTOMOD_INACTIVITY, CHANNEL_AUTOMOD_INACTIVITY

PROD_GUILD_ID = 1172230157776466050
LOGGING_CHANNEL_ID = 44

LOCAL_ENV = [
    "INACTIVE_FORUM_CHANNEL_ID",
    "INACTIVE_FORUM_DURATION",
    "PF_RECRUITMENT_CHANNEL_ID",
    "PF_RECRUITMENT_CHANNEL_DURATION",
    "NOTIFY_CHANNEL_ID",
    "MOD_LOGGING_CHANNEL_ID",
    "GUILD_ID",
    "POSTGRES_HOST",
    "POSTGRES_PORT",
    "POSTGRES_USER",
    "POSTGRES_PASSWORD",
    "POSTGRES_DB",
    "STICKY_ROLE_ARRAY",
    "ANNOUNCEMENT_DRAFT_CHANNEL",
    "MODDINGWAY_ENVIRONMENT",
]


@pytest.fixture
def clean_env(monkeypatch):
    for name in LOCAL_ENV:
        monkeypatch.delenv(name, raising=False)
    return monkeypatch


def test_local__defaults_without_environment(clean_env):
    result = settings.local()

    assert result.model_dump(exclude={"discord_token"}) == {
        "guild_id": 0,
        "log_level": logging.DEBUG,
        "logging_channel_id": 0,
        "notify_channel_id": 0,
        "postgres_host": "localhost",
        "postgres_port": "5432",
        "database_name": "moddingway",
        "postgres_username": "",
        "postgres_password": "",
        "automod_inactivity": {},
        "channel_automod_inactivity": {},
        "sticky_roles": [],
        "announcement_draft_channel": 0,
    }


def test_local__reads_environment(clean_env):
    for name, value in {
        "INACTIVE_FORUM_CHANNEL_ID": "11",
        "INACTIVE_FORUM_DURATION": "30",
        "PF_RECRUITMENT_CHANNEL_ID": "22",
        "PF_RECRUITMENT_CHANNEL_DURATION": "60",
        "NOTIFY_CHANNEL_ID": "33",
        "MOD_LOGGING_CHANNEL_ID": "44",
        "GUILD_ID": "55",
        "POSTGRES_HOST": "db",
        "POSTGRES_PORT": "6543",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "secret",
        "POSTGRES_DB": "testdb",
        "STICKY_ROLE_ARRAY": "[1, 2]",
        "ANNOUNCEMENT_DRAFT_CHANNEL": "66",
    }.items():
        clean_env.setenv(name, value)

    result = settings.local()

    assert result.model_dump(exclude={"discord_token"}) == {
        "guild_id": 55,
        "log_level": logging.DEBUG,
        "logging_channel_id": 44,
        "notify_channel_id": 33,
        "postgres_host": "db",
        "postgres_port": "6543",
        "database_name": "testdb",
        "postgres_username": "user",
        "postgres_password": "secret",
        "automod_inactivity": {11: 30},
        "channel_automod_inactivity": {22: 60},
        "sticky_roles": [1, 2],
        "announcement_draft_channel": 66,
    }


@pytest.mark.parametrize(
    "channel_var,duration_var,field",
    [
        ("INACTIVE_FORUM_CHANNEL_ID", "INACTIVE_FORUM_DURATION", "automod_inactivity"),
        (
            "PF_RECRUITMENT_CHANNEL_ID",
            "PF_RECRUITMENT_CHANNEL_DURATION",
            "channel_automod_inactivity",
        ),
    ],
)
def test_local__inactivity_needs_channel_and_duration(
    clean_env, channel_var, duration_var, field
):
    clean_env.setenv(channel_var, "11")
    assert getattr(settings.local(), field) == {}

    clean_env.delenv(channel_var)
    clean_env.setenv(duration_var, "30")
    assert getattr(settings.local(), field) == {}


def test_local__notify_channel_falls_back_to_logging_channel(clean_env):
    clean_env.setenv("MOD_LOGGING_CHANNEL_ID", str(LOGGING_CHANNEL_ID))

    assert settings.local().notify_channel_id == LOGGING_CHANNEL_ID


@pytest.mark.parametrize("value", ["", "abc"])
def test_local__blank_or_invalid_ids_fall_back_to_zero(clean_env, value):
    clean_env.setenv("GUILD_ID", value)

    assert settings.local().guild_id == 0


@pytest.mark.parametrize("environment", ["prod", "PROD"])
def test_get_settings__prod(clean_env, environment):
    clean_env.setenv("MODDINGWAY_ENVIRONMENT", environment)
    clean_env.setenv("POSTGRES_HOST", "prod-db")
    clean_env.setenv("POSTGRES_PORT", "5432")

    result = settings.get_settings()

    assert result.guild_id == PROD_GUILD_ID
    assert result.log_level == logging.INFO
    assert result.postgres_host == "prod-db"
    assert result.postgres_port == "5432"
    assert result.automod_inactivity == AUTOMOD_INACTIVITY
    assert result.channel_automod_inactivity == CHANNEL_AUTOMOD_INACTIVITY


@pytest.mark.parametrize("environment", [None, "local", "dev"])
def test_get_settings__local(clean_env, environment):
    if environment is not None:
        clean_env.setenv("MODDINGWAY_ENVIRONMENT", environment)

    assert settings.get_settings().log_level == logging.DEBUG
