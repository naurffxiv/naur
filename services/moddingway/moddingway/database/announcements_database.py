import json
import logging
from typing import Any

from moddingway.settings import get_settings

from . import DatabaseConnection
from .models import AnnouncementRevision

settings = get_settings()

logger = logging.getLogger(__name__)


def add_announcement(announcement_rev: AnnouncementRevision) -> int:
    conn = DatabaseConnection()

    revisions_data = [announcement_rev.model_dump()]

    sent_flag = False
    discord_msg_id = None

    with conn.get_cursor() as cursor:
        query = """
            INSERT INTO announcements
            (announcementRevisions, sentFLAG, discordMessageID)
            VALUES
            (%s,%s,%s)
            RETURNING announcementID
        """

        params = (json.dumps(revisions_data), sent_flag, discord_msg_id)

        cursor.execute(query, params)
        res = cursor.fetchone()

        if res is None:
            raise ValueError("Failed to add announcement to DB")

        return res[0]


def get_announcement(announcement_id: int) -> dict[str, Any] | None:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
            SELECT announcementID, announcementRevisions, sentFLAG, discordMessageID
            FROM announcements
            WHERE announcementID = %s
        """

        cursor.execute(query, (announcement_id,))
        res = cursor.fetchone()

        if res is None:
            return None

        announcement_id = res[0]
        revisions_data = res[1]
        sent_flag = res[2]
        discord_msg_id = res[3]

        if isinstance(revisions_data, str):
            revisions_data = json.loads(revisions_data)

        return {
            "announcement_id": announcement_id,
            "revisions": revisions_data,
            "sent_flag": sent_flag,
            "discord_msg_id": discord_msg_id,
        }


def set_sent(announcement_id: int, discord_msg_id: int) -> bool:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
            UPDATE announcements
            SET
                discordMessageID = %s,
                sentFLAG = %s
            WHERE announcementID = %s
        """

        # Setting sent_flag to True and linking it to the announcement messages id
        params = (discord_msg_id, True, announcement_id)

        cursor.execute(query, params)

        if cursor.rowcount == 0:
            raise ValueError(
                f"Failed to update: Announcement ID {announcement_id} not found."
            )

        return True
