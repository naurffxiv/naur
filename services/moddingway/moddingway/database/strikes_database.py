from dataclasses import dataclass
from datetime import datetime

from moddingway.constants import StrikeSeverity

from . import DatabaseConnection
from .models import Strike


@dataclass
class StrikeDisplay:
    strike_id: int
    user_id: int
    discord_user_id: str
    severity: StrikeSeverity
    reason: str
    created_by: str
    created_timestamp: datetime
    last_edited_by: str
    last_edited_timestamp: datetime


def _convert_row_to_strike_display(row: tuple) -> StrikeDisplay:
    try:
        return StrikeDisplay(
            strike_id=row[0],
            user_id=row[1],
            discord_user_id=row[2],
            severity=StrikeSeverity(row[3]),
            reason=row[4],
            created_by=row[5],
            created_timestamp=row[6],
            last_edited_by=row[7],
            last_edited_timestamp=row[8],
        )
    except IndexError as e:
        raise ValueError(
            "Invalid row data encountered while converting to StrikeDisplay"
        ) from e


def add_strike(strike: Strike) -> int:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
            INSERT INTO strikes
            (userID, severity, reason, createdTimestamp, createdBy, lastEditedTimestamp, lastEditedBy)
            VALUES
            (%s, %s, %s, %s, %s, %s, %s)
            RETURNING strikeId
        """

        params = (
            strike.user_id,
            strike.severity,
            strike.reason,
            strike.created_timestamp,
            strike.created_by,
            strike.last_edited_timestamp,
            strike.last_edited_by,
        )

        cursor.execute(query, params)
        res = cursor.fetchone()

        if res is None:
            raise ValueError("Failed to add strike to DB")

        return res[0]


def list_strikes(user_id: int) -> list[tuple]:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
        select s.strikeid, s.severity, s.reason, s.createdby, s.createdtimestamp
        from strikes s
        join users u on u.userID = s.userID
        where u.userId = %s
        order by s.createdtimestamp asc
        """

        params = (user_id,)

        cursor.execute(query, params)
        res = cursor.fetchall()

        return res


def delete_strike(strike_id: int) -> tuple[str, str, str] | None:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
        delete from strikes s where s.strikeId = %s
        returning s.strikeid, s.userId, s.severity
        """

        params = (strike_id,)

        cursor.execute(query, params)
        res = cursor.fetchall()

        if len(res) == 0:
            return None

        return res[0]


def get_strike(strike_id: int) -> StrikeDisplay | None:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
        select s.strikeid, s.userid, u.discorduserid, s.severity, s.reason,
               s.createdby, s.createdtimestamp, s.lasteditedby, s.lasteditedtimestamp
        from strikes s
        join users u on u.userid = s.userid
        where s.strikeid = %s
        """

        params = (strike_id,)

        cursor.execute(query, params)
        row = cursor.fetchone()

        if row is None:
            return None

        return _convert_row_to_strike_display(row)


def update_strike(
    strike_id: int,
    reason: str,
    severity: StrikeSeverity,
    editor_id: str,
    update_timestamp: datetime,
) -> bool:
    conn = DatabaseConnection()

    with conn.get_cursor() as cursor:
        query = """
        update strikes s
        set reason = %s,
            severity = %s,
            createdby = %s,
            lasteditedby = %s,
            lasteditedtimestamp = %s
        where s.strikeid = %s
        """

        params = (
            reason,
            severity,
            editor_id,
            editor_id,
            update_timestamp,
            strike_id,
        )

        cursor.execute(query, params)
        return cursor.rowcount == 1
