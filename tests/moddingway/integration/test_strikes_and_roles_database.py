from datetime import datetime

from moddingway.constants import StrikeSeverity
from moddingway.database import roles_database, strikes_database, users_database
from moddingway.database.models import Strike


def _strike(user_id, when, severity=StrikeSeverity.MINOR, reason="spam"):
    return strikes_database.add_strike(
        Strike(
            user_id=user_id,
            severity=severity,
            reason=reason,
            created_timestamp=when,
            created_by="99",
            last_edited_timestamp=when,
            last_edited_by="99",
        )
    )


def test_list_strikes_in_creation_order():
    user = users_database.add_user(1)
    later = _strike(user.user_id, datetime(2026, 2, 1), StrikeSeverity.SERIOUS, "later")
    earlier = _strike(
        user.user_id, datetime(2026, 1, 1), StrikeSeverity.MINOR, "earlier"
    )

    strikes = strikes_database.list_strikes(user.user_id)

    assert strikes == [
        (earlier, StrikeSeverity.MINOR, "earlier", "99", datetime(2026, 1, 1)),
        (later, StrikeSeverity.SERIOUS, "later", "99", datetime(2026, 2, 1)),
    ]


def test_delete_strike_returns_what_was_deleted():
    user = users_database.add_user(1)
    strike_id = _strike(user.user_id, datetime(2026, 1, 1), StrikeSeverity.MODERATE)

    assert strikes_database.delete_strike(strike_id) == (
        strike_id,
        user.user_id,
        StrikeSeverity.MODERATE,
    )
    assert strikes_database.delete_strike(strike_id) is None
    assert strikes_database.list_strikes(user.user_id) == []


def test_sticky_roles_round_trip():
    user = users_database.add_user(1)

    roles_database.add_sticky_roles(user.user_id, [])
    assert roles_database.get_sticky_roles(user.user_id) == []

    roles_database.add_sticky_roles(user.user_id, [111, 222])
    assert sorted(roles_database.get_sticky_roles(user.user_id)) == ["111", "222"]

    removed = roles_database.remove_sticky_roles(user.user_id)
    assert sorted(r[0] for r in removed) == ["111", "222"]
    assert roles_database.get_sticky_roles(user.user_id) == []
