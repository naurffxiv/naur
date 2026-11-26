from datetime import datetime, timedelta

from moddingway.constants import UserRole
from moddingway.database import users_database


def _set_role(database, user_id, role):
    with database.get_cursor() as cursor:
        cursor.execute(
            "UPDATE users SET userRole = %s WHERE userID = %s", (role, user_id)
        )


def test_add_and_get_user():
    created = users_database.add_user(12345)

    fetched = users_database.get_user(12345)

    assert fetched == created
    assert fetched.discord_user_id == "12345"
    assert fetched.user_role == UserRole.USER
    assert fetched.get_strike_points() == 0
    assert users_database.get_user(999) is None


def test_update_user_strike_points():
    user = users_database.add_user(1)
    when = datetime(2026, 1, 2, 3, 4, 5)
    user.temporary_points, user.permanent_points, user.last_infraction_timestamp = (
        3,
        7,
        when,
    )

    users_database.update_user_strike_points(user)

    fetched = users_database.get_user(1)
    assert (fetched.temporary_points, fetched.permanent_points) == (3, 7)
    assert fetched.last_infraction_timestamp == when


def test_decrement_user_strike_points_never_goes_below_zero():
    user = users_database.add_user(1)
    user.temporary_points, user.permanent_points = 2, 7
    users_database.update_user_strike_points(user)

    users_database.decrement_user_strike_points(user.user_id, 5, 3)

    fetched = users_database.get_user(1)
    assert (fetched.temporary_points, fetched.permanent_points) == (0, 4)


def test_decrement_old_strike_points_only_touches_old_unbanned_infractions():
    old = datetime.now() - timedelta(days=100)
    recent = datetime.now() - timedelta(days=10)
    for discord_id, points, last, banned in [
        (1, 3, old, False),
        (2, 3, recent, False),
        (3, 3, old, True),
        (4, 0, old, False),
    ]:
        user = users_database.add_user(discord_id)
        user.temporary_points, user.last_infraction_timestamp, user.is_banned = (
            points,
            last,
            banned,
        )
        users_database.update_user(user)

    assert users_database.decrement_old_strike_points() == 1

    decremented = users_database.get_user(1)
    assert decremented.last_infraction_timestamp == old + timedelta(days=90)
    points = [users_database.get_user(i).temporary_points for i in (1, 2, 3, 4)]
    assert points == [2, 3, 3, 0]


USER_COUNT = 5


def test_counts_and_pages(database):
    for discord_id in range(1, USER_COUNT + 1):
        users_database.add_user(discord_id)
    mod = users_database.get_user(2)
    _set_role(database, mod.user_id, UserRole.MOD)
    banned = users_database.get_user(3)
    banned.is_banned = True
    users_database.update_user(banned)

    assert users_database.get_user_count() == USER_COUNT
    assert [u.discord_user_id for u in users_database.get_users(2, 1)] == ["2", "3"]
    assert users_database.get_users(10, 10) == []
    assert users_database.get_mod_count() == 1
    assert [u.discord_user_id for u in users_database.get_mods(10, 0)] == ["2"]
    assert users_database.get_banned_count() == 1
    assert [u.discord_user_id for u in users_database.get_banned_users(10, 0)] == ["3"]
    assert users_database.get_banned_users(10, 1) == []
