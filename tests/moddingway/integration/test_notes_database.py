from datetime import datetime

import psycopg2
import pytest

from moddingway.database import notes_database, users_database
from moddingway.database.models import Note


def _note(user_id, text, when, warning=False):
    return notes_database.add_note(
        Note(
            user_id=user_id,
            is_warning=warning,
            note=text,
            created_timestamp=when,
            created_by="1",
            last_edited_timestamp=when,
            last_edited_by="1",
        )
    )


def test_list_notes_and_warnings():
    user = users_database.add_user(1)
    second = _note(user.user_id, "second", datetime(2026, 2, 1), warning=True)
    first = _note(user.user_id, "first", datetime(2026, 1, 1))

    assert [n.note_id for n in notes_database.list_notes(user.user_id)] == [
        first,
        second,
    ]
    warnings = notes_database.list_warnings(user.user_id)
    assert [(n.note_id, n.is_warning, n.content) for n in warnings] == [
        (second, True, "second")
    ]


def test_update_and_delete_note():
    user = users_database.add_user(1)
    note_id = _note(user.user_id, "old", datetime(2026, 1, 1))
    edited_at = datetime(2026, 1, 5)

    assert notes_database.update_note("new", "42", edited_at, note_id) is True
    note = notes_database.get_note(note_id)
    assert (
        note.content,
        note.last_editor,
        note.created_by,
        note.last_edited_timestamp,
    ) == ("new", "42", "1", edited_at)

    assert notes_database.delete_note(note_id) is True
    assert notes_database.delete_note(note_id) is False
    assert notes_database.update_note("x", "42", edited_at, note_id) is False


def test_get_note_returns_none_when_missing():
    # note_service treats a missing note as None ("Note not found")
    assert notes_database.get_note(999) is None


def test_notes_are_limited_to_300_characters():
    user = users_database.add_user(1)

    with pytest.raises(psycopg2.errors.CheckViolation):
        _note(user.user_id, "x" * 301, datetime(2026, 1, 1))
