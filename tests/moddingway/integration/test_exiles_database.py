from datetime import UTC, datetime, timedelta

from moddingway.constants import ExileStatus
from moddingway.database import exiles_database, users_database
from moddingway.database.models import Exile


def _exile(
    user_id, status=ExileStatus.TIMED_EXILED, start=None, end=None, reason="spam"
):
    start = start or datetime(2026, 1, 1)
    return exiles_database.add_exile(
        Exile(
            None, user_id, None, reason, status, start, end or start + timedelta(days=1)
        )
    )


def test_active_exile_lifecycle():
    user = users_database.add_user(12345)
    exile_id = _exile(user.user_id)

    active = exiles_database.get_user_active_exile(user.user_id)
    assert (active.exile_id, active.user_id, active.discord_id) == (
        exile_id,
        user.user_id,
        12345,
    )
    assert exiles_database.get_exile_status(exile_id) == ExileStatus.TIMED_EXILED

    new_end = datetime(2026, 2, 1)
    exiles_database.update_exile_end(exile_id, new_end)
    assert exiles_database.get_user_active_exile(user.user_id).end_timestamp == new_end

    exiles_database.update_exile_status(exile_id, ExileStatus.UNEXILED)
    assert exiles_database.get_user_active_exile(user.user_id) is None
    assert exiles_database.get_exile_status(exile_id) == ExileStatus.UNEXILED


def test_get_pending_unexiles_returns_only_expired_active_exiles():
    user = users_database.add_user(1)
    now = datetime.now(UTC).replace(tzinfo=None)
    expired = _exile(
        user.user_id, start=now - timedelta(days=2), end=now - timedelta(minutes=1)
    )
    _exile(user.user_id, start=now, end=now + timedelta(days=1))
    _exile(
        user.user_id,
        status=ExileStatus.UNEXILED,
        start=now - timedelta(days=2),
        end=now - timedelta(days=1),
    )

    pending = exiles_database.get_pending_unexiles()

    assert [p.exile_id for p in pending] == [expired]
    assert pending[0].discord_id == 1


def test_user_exiles_are_ordered_and_all_active_exiles_listed():
    a = users_database.add_user(1)
    b = users_database.add_user(2)
    later = _exile(
        a.user_id,
        status=ExileStatus.UNEXILED,
        start=datetime(2026, 3, 1),
        reason="later",
    )
    earlier = _exile(a.user_id, start=datetime(2026, 1, 1), reason="earlier")
    other = _exile(b.user_id)

    history = exiles_database.get_user_exiles(a.user_id)
    assert [(e.exile_id, e.reason, e.discord_id) for e in history] == [
        (earlier, "earlier", "1"),
        (later, "later", "1"),
    ]
    assert sorted(
        e.exile_id for e in exiles_database.get_all_active_exiles()
    ) == sorted([earlier, other])
    assert exiles_database.get_user_exiles(999) == []


def test_delete_exile_refuses_active_exiles():
    user = users_database.add_user(1)
    active = _exile(user.user_id)
    finished = _exile(user.user_id, status=ExileStatus.UNEXILED)

    assert exiles_database.delete_exile(active) is False
    assert exiles_database.delete_exile(finished) is True
    assert exiles_database.get_exile_status(finished) is None
    assert exiles_database.delete_exile(finished) is False


def test_remove_user_exiles():
    user = users_database.add_user(1)
    assert exiles_database.remove_user_exiles(user.user_id) is None
    _exile(user.user_id)

    assert exiles_database.remove_user_exiles(user.user_id) is not None
    assert exiles_database.get_user_exiles(user.user_id) == []
