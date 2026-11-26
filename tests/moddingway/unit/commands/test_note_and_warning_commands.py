import pytest
from command_helpers import logged, reply, run, sent, target
from fakes import embed_fields
from pytest_mock.plugin import MockerFixture

from moddingway.commands.note_commands import NoteDeleteView, create_note_commands
from moddingway.commands.warning_commands import create_warning_commands


@pytest.fixture
def note_bot(bot):
    create_note_commands(bot)
    create_warning_commands(bot)
    return bot


@pytest.mark.asyncio
async def test_add_note_records_note_and_logs_it(
    note_bot, make_interaction, mocker: MockerFixture
):
    add_note = mocker.patch("moddingway.services.note_service.add_note")
    i = make_interaction("add_note")
    user = target(mocker)

    await run(note_bot, "add_note", i, user=user, note="be careful")

    add_note.assert_awaited_once()
    assert add_note.await_args.args[1:] == (user, "be careful", i.user)
    assert reply(i) == "Successfully added note to <@12345>"
    embed = logged(i)
    assert embed.description == "Used `add_note` command in #mod-chat"
    assert {("Action", "/add_note"), ("Note", "be careful")} <= set(embed_fields(embed))


@pytest.mark.asyncio
async def test_warn_adds_a_warning_note(
    note_bot, make_interaction, mocker: MockerFixture
):
    add_note = mocker.patch("moddingway.services.note_service.add_note")
    i = make_interaction("warn")

    await run(note_bot, "warn", i, user=target(mocker), warning_note="stop")

    assert add_note.await_args.args[-1] is True
    assert reply(i) == "Successfully added warning to <@12345>"


@pytest.mark.parametrize(
    "command,service",
    [("view_notes", "get_user_notes"), ("view_warnings", "get_user_warnings")],
)
@pytest.mark.asyncio
async def test_view_commands_reply_with_service_text(
    note_bot, make_interaction, mocker: MockerFixture, command, service
):
    mocker.patch(f"moddingway.services.note_service.{service}", return_value="the list")
    i = make_interaction(command)

    await run(note_bot, command, i, user=target(mocker))

    assert reply(i) == "the list"


@pytest.mark.asyncio
async def test_delete_note_asks_for_confirmation(
    note_bot, make_interaction, mocker: MockerFixture
):
    mocker.patch(
        "moddingway.services.note_service.get_note_by_id",
        return_value="\n* ID: 7 | Note: x",
    )
    i = make_interaction("delete_note")

    await run(note_bot, "delete_note", i, note_id=7)

    assert sent(i) == "Are you sure you want to delete this note? \n\n* ID: 7 | Note: x"
    view = i.response.send_message.await_args.kwargs["view"]
    assert isinstance(view, NoteDeleteView)
    assert view.note_id == 7  # noqa: PLR2004 - the note_id passed above


@pytest.mark.asyncio
async def test_delete_note_for_missing_note_does_not_ask_for_confirmation(
    note_bot, make_interaction, mocker: MockerFixture
):
    get_note = mocker.patch(
        "moddingway.database.notes_database.get_note", return_value=None
    )
    i = make_interaction("delete_note")

    await run(note_bot, "delete_note", i, note_id=7)

    get_note.assert_called_once_with(7)
    assert sent(i) == "Note not found"
    assert "view" not in i.response.send_message.await_args.kwargs


@pytest.mark.asyncio
async def test_note_delete_view_confirm_and_cancel(
    make_interaction, mocker: MockerFixture
):
    delete = mocker.patch(
        "moddingway.services.note_service.delete_user_note",
        return_value="Successfully deleted note: 7",
    )
    original = make_interaction("delete_note")
    button = make_interaction()

    view = NoteDeleteView(note_id=7, interaction=original)
    await view.confirm_button_callback.callback(button)

    delete.assert_awaited_once()
    assert reply(button) == "Successfully deleted note: 7"
    original.edit_original_response.assert_awaited_once_with(view=view)
    assert view.children == []

    cancel = make_interaction()
    await NoteDeleteView(
        note_id=7, interaction=original
    ).cancel_button_callback.callback(cancel)
    assert reply(cancel) == "Note deletion cancelled"


@pytest.mark.asyncio
async def test_edit_note(note_bot, make_interaction, mocker: MockerFixture):
    update = mocker.patch(
        "moddingway.services.note_service.update_user_note",
        return_value="Note succesfully updated",
    )
    i = make_interaction("edit_note")

    await run(note_bot, "edit_note", i, new_note="fixed", note_id=7)

    assert update.await_args.args[1:] == (i.user, "fixed", 7)
    assert reply(i) == "Note succesfully updated"
