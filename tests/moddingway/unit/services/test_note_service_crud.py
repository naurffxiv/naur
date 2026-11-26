import datetime

import discord
import pytest
from fakes import embed_fields
from pytest_mock.plugin import MockerFixture

from moddingway.database.notes_database import NoteDisplay
from moddingway.services import note_service

EDITED = datetime.datetime(2024, 6, 15, 12, 0, 0, tzinfo=datetime.UTC)
EDITED_EPOCH = int(EDITED.timestamp())


def _note(note_id=1, warning=False, content="text"):
    return NoteDisplay(
        note_id=note_id,
        is_warning=warning,
        content=content,
        created_by=2,
        last_editor=3,
        last_edited_timestamp=EDITED,
    )


@pytest.mark.parametrize(
    "warning,dm",
    [
        (False, None),
        (
            True,
            "You have been issued a warning from NA Ultimate Raiding - FFXIV:\nwatch out",
        ),
    ],
)
@pytest.mark.asyncio
async def test_add_note(
    mocker: MockerFixture, create_member, create_db_user, warning, dm
):
    mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=create_db_user(user_id=1),
    )
    add = mocker.patch("moddingway.database.notes_database.add_note", return_value=33)
    member = create_member(id=12345)
    embed = discord.Embed()

    await note_service.add_note(
        embed, member, "watch out", mocker.Mock(id=2), is_warning=warning
    )

    note = add.call_args.args[0]
    assert (
        note.user_id,
        note.is_warning,
        note.note,
        note.created_by,
        note.last_edited_by,
    ) == (1, warning, "watch out", "2", "2")
    assert embed.footer.text == "Note ID: 33"
    assert embed_fields(embed)[-1] == ("Result", "<@12345> was given a note")
    send = member.create_dm.return_value.send
    assert send.await_args_list == ([mocker.call(content=dm)] if dm else [])


@pytest.mark.asyncio
async def test_add_note_creates_missing_user(
    mocker: MockerFixture, create_member, create_db_user
):
    mocker.patch("moddingway.database.users_database.get_user", return_value=None)
    add_user = mocker.patch(
        "moddingway.database.users_database.add_user",
        return_value=create_db_user(user_id=4),
    )
    add = mocker.patch("moddingway.database.notes_database.add_note", return_value=1)
    embed = discord.Embed()

    await note_service.add_note(embed, create_member(id=12345), "x", mocker.Mock(id=2))

    add_user.assert_called_once_with(12345)
    assert add.call_args.args[0].user_id == add_user.return_value.user_id
    assert embed.description == "\nUser not found in database, creating new record"


@pytest.mark.parametrize("warning,prefix", [(False, ""), (True, "[warning] ")])
@pytest.mark.asyncio
async def test_get_note_by_id(mocker: MockerFixture, warning, prefix):
    mocker.patch(
        "moddingway.database.notes_database.get_note",
        return_value=_note(warning=warning),
    )

    assert await note_service.get_note_by_id(1) == (
        f"\n* ID: 1 | Note: {prefix}text | Note Creator: <@2> | Last Editor: <@3> | Last Edited: <t:{EDITED_EPOCH}:f>"
    )


@pytest.mark.asyncio
async def test_get_user_warnings(mocker: MockerFixture, create_db_user):
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user", return_value=None
    )
    list_warnings = mocker.patch(
        "moddingway.database.notes_database.list_warnings", return_value=[]
    )
    user = mocker.Mock(id=12345)

    assert await note_service.get_user_warnings(user) == "User not found in database"
    get_user.return_value = create_db_user(user_id=1)
    assert (
        await note_service.get_user_warnings(user)
        == "No warnings found for user <@12345>"
    )
    list_warnings.return_value = [_note(5, warning=True, content="be nice")]
    assert await note_service.get_user_warnings(user) == (
        "Warnings found for <@12345>:"
        f"\n* ID: 5 | Note Creator: <@2> | Last Editor: <@3> | Last Edited: <t:{EDITED_EPOCH}:f> | Note: [warning] be nice"
    )
    list_warnings.assert_called_with(1)


@pytest.mark.parametrize(
    "note,deleted,expected,result_field,footer",
    [
        (
            None,
            None,
            "Note not found in database, no action will be taken",
            "Note not found, no action will be taken",
            None,
        ),
        (_note(), True, "Successfully deleted note: 7", "Note deleted", "Note ID: 7"),
        (
            _note(),
            False,
            "There was an error deleting note from the database",
            "Error deleting note",
            None,
        ),
    ],
)
@pytest.mark.asyncio
async def test_delete_user_note(
    mocker: MockerFixture, note, deleted, expected, result_field, footer
):
    mocker.patch("moddingway.database.notes_database.get_note", return_value=note)
    delete = mocker.patch(
        "moddingway.database.notes_database.delete_note", return_value=deleted
    )
    embed = discord.Embed()

    assert await note_service.delete_user_note(embed, 7) == expected

    assert delete.called == (note is not None)
    assert embed_fields(embed) == [("Result", result_field)]
    assert embed.footer.text == footer


@pytest.mark.asyncio
async def test_update_user_note(mocker: MockerFixture):
    get_note = mocker.patch(
        "moddingway.database.notes_database.get_note", return_value=_note(content="old")
    )
    update = mocker.patch(
        "moddingway.database.notes_database.update_note", return_value=True
    )
    embed = discord.Embed()

    assert (
        await note_service.update_user_note(embed, mocker.Mock(id=9), "new", 7)
        == "Note succesfully updated"
    )

    new_note, author, _when, note_id = update.call_args.args
    assert (new_note, author, note_id) == ("new", "9", 7)
    assert embed.footer.text == "Note ID: 7"
    assert embed_fields(embed) == [("Old note", "old"), ("Result", "Note updated")]

    update.return_value = False
    assert (
        await note_service.update_user_note(
            discord.Embed(), mocker.Mock(id=9), "new", 7
        )
        == "There was an error updating the note"
    )

    get_note.return_value = None
    embed = discord.Embed()
    assert (
        await note_service.update_user_note(embed, mocker.Mock(id=9), "new", 7)
        == "Note not found in database"
    )
    assert embed_fields(embed) == [("Result", "Note not found")]
