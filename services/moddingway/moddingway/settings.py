import logging
import os
from ast import literal_eval

from dotenv import load_dotenv
from pydantic import BaseModel

from .constants import AUTOMOD_INACTIVITY, CHANNEL_AUTOMOD_INACTIVITY, STICKY_ROLES

load_dotenv()

logger = logging.getLogger(__name__)


class Settings(BaseModel):
    """Class for keeping track of settings"""

    guild_id: int
    discord_token: str = os.environ.get("DISCORD_TOKEN", "")
    log_level: int = logging.INFO
    logging_channel_id: int
    notify_channel_id: int
    postgres_host: str
    postgres_port: str
    database_name: str = "moddingway"
    postgres_username: str = os.environ.get("POSTGRES_USER", "")
    postgres_password: str = os.environ.get("POSTGRES_PASSWORD", "")
    automod_inactivity: dict[int, int]  # key: channel id, value: inactive limit (days)
    channel_automod_inactivity: dict[
        int, int
    ]  # key: channel id, value: inactive limit (minutes)
    sticky_roles: list[
        int
    ]  # roles that grant access to channels that should be stripped/restored on exile/unexile


def prod() -> Settings:
    return Settings(
        guild_id=1172230157776466050,
        logging_channel_id=1172324840947056681,  # mod-reports
        notify_channel_id=1279952544235524269,  # bot-channel
        log_level=logging.INFO,
        postgres_host=os.environ.get("POSTGRES_HOST", ""),
        postgres_port=os.environ.get("POSTGRES_PORT", ""),
        automod_inactivity=AUTOMOD_INACTIVITY,
        channel_automod_inactivity=CHANNEL_AUTOMOD_INACTIVITY,
        sticky_roles=STICKY_ROLES,
    )


def local() -> Settings:
    automod_inactivity = {}
    channel_automod_inactivity = {}

    inactive_forum_channel_id = os.environ.get("INACTIVE_FORUM_CHANNEL_ID", "")
    inactive_forum_duration = os.environ.get("INACTIVE_FORUM_DURATION", "")
    if inactive_forum_channel_id != "" and inactive_forum_duration != "":
        automod_inactivity[int(inactive_forum_channel_id)] = int(
            inactive_forum_duration
        )

    pf_recruitment_channel_id = os.environ.get("PF_RECRUITMENT_CHANNEL_ID", "")
    pf_recruitment_channel_duration = os.environ.get(
        "PF_RECRUITMENT_CHANNEL_DURATION", ""
    )
    if pf_recruitment_channel_id != "" and pf_recruitment_channel_duration != "":
        channel_automod_inactivity[int(pf_recruitment_channel_id)] = int(
            pf_recruitment_channel_duration
        )

    notify_channel_id_str = os.environ.get("NOTIFY_CHANNEL_ID") or os.environ.get(
        "MOD_LOGGING_CHANNEL_ID"
    )
    if not notify_channel_id_str:
        notify_channel_id_str = "0"

    return Settings(
        guild_id=_check_env_convert("GUILD_ID"),
        logging_channel_id=_check_env_convert("MOD_LOGGING_CHANNEL_ID"),
        log_level=logging.DEBUG,
        postgres_host=os.environ.get("POSTGRES_HOST", "localhost"),
        postgres_port=os.environ.get("POSTGRES_PORT", "5432"),
        automod_inactivity=automod_inactivity,
        channel_automod_inactivity=channel_automod_inactivity,
        notify_channel_id=int(notify_channel_id_str),
        postgres_username=os.environ.get("POSTGRES_USER", ""),
        postgres_password=os.environ.get("POSTGRES_PASSWORD", ""),
        database_name=os.environ.get("POSTGRES_DB", "moddingway"),
        sticky_roles=literal_eval(os.environ.get("STICKY_ROLE_ARRAY") or "None") or [],
    )


def get_settings() -> Settings:
    try:
        env_name = os.environ["MODDINGWAY_ENVIRONMENT"].lower()
    except KeyError:
        env_name = "local"
    if env_name == "prod":
        return prod()
    else:
        return local()


def _check_env_convert(key: str, default: int = 0) -> int:
    value_str = os.environ.get(key)

    if value_str is None or value_str == "":
        return default

    try:
        return int(value_str)
    except ValueError:
        logger.info(f"ENV {key} is not set to {value_str}, returning {default}.")
        return default
