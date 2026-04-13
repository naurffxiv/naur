import json
import logging
from typing import Any

from moddingway.constants import ANNOUNCEMENT_REV_LIMIT
from moddingway.settings import get_settings

from . import DatabaseConnection
from .models import AnnouncementRevision

settings = get_settings()

logger = logging.getLogger(__name__)


def insert_announcement(announcement_rev: AnnouncementRevision) -> int:
    conn = DatabaseConnection()

    revisions_data = [announcement_rev.model_dump()]

    sent_flag = False
    discord_msg_link = None

    with conn.get_cursor() as cursor:
        query = """
            INSERT INTO announcements
            (announcementRevisions, sentFLAG, discordMessageLink)
            VALUES
            (%s,%s,%s)
            RETURNING announcementID
        """

        params = (json.dumps(revisions_data), sent_flag, discord_msg_link)

        cursor.execute(query, params)
        res = cursor.fetchone()

        if res is None:
            raise ValueError("Failed to add announcement to DB")

        return res[0]


def get_announcement(announcement_id: int) -> dict[str, Any] | None:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
            SELECT announcementID, announcementRevisions, sentFLAG, discordMessageLink
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
        discord_msg_link = res[3]

        if isinstance(revisions_data, str):
            revisions_data = json.loads(revisions_data)

        return {
            "announcement_id": announcement_id,
            "revisions": revisions_data,
            "sent_flag": sent_flag,
            "discord_msg_link": discord_msg_link,
        }


def set_sent(announcement_id: int, discord_msg_link: str) -> bool:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
            UPDATE announcements
            SET
                discordMessageLink = %s,
                sentFLAG = %s
            WHERE announcementID = %s
        """

        # Setting sent_flag to True and linking it to the announcement channel id and message id
        params = (discord_msg_link, True, announcement_id)

        cursor.execute(query, params)

        if cursor.rowcount == 0:
            raise ValueError(
                f"Failed to update: Announcement ID {announcement_id} not found."
            )

        return True


def select_announcements_bulk(status: bool | None = None):
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
        SELECT announcementID, announcementRevisions, sentFLAG, discordMessageLink
        FROM announcements
        WHERE sentFLAG = %s OR %s IS NULL
        """

        params = (status, status)

        cursor.execute(query, params)

        res = cursor.fetchall()

        return res


def add_revision(announcement_id: int, author_id: int, content: str) -> bool:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        fetch_query = """
            SELECT announcementRevisions
            FROM announcements
            WHERE announcementID = %s
        """
        cursor.execute(fetch_query, (announcement_id,))
        res = cursor.fetchone()

        if res is None:
            raise ValueError(f"Announcement ID {announcement_id} not found.")

        revisions_data = res[0]

        if isinstance(revisions_data, str):
            revisions_data = json.loads(revisions_data)

        ### Calculate next ver
        if revisions_data and "version" in revisions_data[-1]:
            next_version = revisions_data[-1]["version"] + 1
        else:
            ### fallback for old data or empty lists
            next_version = len(revisions_data) + 1

        revisions_data.append(
            {"version": next_version, "author_id": author_id, "content": content}
        )

        ### Keep it three revs
        if len(revisions_data) > ANNOUNCEMENT_REV_LIMIT:
            revisions_data = revisions_data[-ANNOUNCEMENT_REV_LIMIT:]

        update_query = """
            UPDATE announcements
            SET announcementRevisions = %s
            WHERE announcementID = %s
        """
        cursor.execute(update_query, (json.dumps(revisions_data), announcement_id))

        if cursor.rowcount == 0:
            raise ValueError(
                f"Failed to update revisions for Announcement ID {announcement_id}."
            )

        return True
