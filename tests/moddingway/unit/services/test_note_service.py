import datetime

import pytest
from pytest_mock.plugin import MockerFixture

from moddingway.database.notes_database import NoteDisplay
from moddingway.services import note_service

DEFAULT_LAST_EDITED = datetime.datetime(2024, 6, 15, 12, 0, 0)
DEFAULT_LAST_EDITED_EPOCH = int(
    DEFAULT_LAST_EDITED.replace(tzinfo=datetime.UTC).timestamp()
)


@pytest.mark.asyncio
async def test_get_user_notes__plain_note_includes_last_edited_timestamp(
    mocker: MockerFixture, create_member, create_db_user
):
    # Arrange
    user = create_member(id=999)
    note = NoteDisplay(
        note_id=7,
        is_warning=False,
        content="be nice",
        created_by=111,
        last_editor=222,
        last_edited_timestamp=DEFAULT_LAST_EDITED,
    )
    mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=create_db_user(user_id=42),
    )
    mocker.patch(
        "moddingway.database.notes_database.list_notes",
        return_value=[note],
    )

    # Act
    result = await note_service.get_user_notes(user)

    # Assert
    assert "ID: 7" in result
    assert "Note Creator: <@111>" in result
    assert "Last Editor: <@222>" in result
    assert f"Last Edited: <t:{DEFAULT_LAST_EDITED_EPOCH}:f>" in result
    assert "Note: be nice" in result
    assert "[warning]" not in result


@pytest.mark.asyncio
async def test_get_note_by_id__plain_note_includes_last_edited_timestamp(
    mocker: MockerFixture,
):
    # Arrange
    note = NoteDisplay(
        note_id=5,
        is_warning=False,
        content="a plain note",
        created_by=111,
        last_editor=222,
        last_edited_timestamp=DEFAULT_LAST_EDITED,
    )
    mocker.patch(
        "moddingway.database.notes_database.get_note",
        return_value=note,
    )

    # Act
    result = await note_service.get_note_by_id(5)

    # Assert
    assert "ID: 5" in result
    assert "Note: a plain note" in result
    assert "Note Creator: <@111>" in result
    assert "Last Editor: <@222>" in result
    assert f"Last Edited: <t:{DEFAULT_LAST_EDITED_EPOCH}:f>" in result
    assert "[warning]" not in result
