from . import DatabaseConnection
from .models import Strike


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
