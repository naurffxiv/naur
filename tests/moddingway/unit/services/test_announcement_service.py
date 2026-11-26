import discord
import pytest
from fakes import embed_fields
from pytest_mock.plugin import MockerFixture

from moddingway.services import announcement_service


@pytest.mark.asyncio
async def test_add_announcement_stores_a_draft(mocker: MockerFixture):
    insert = mocker.patch(
        "moddingway.database.announcements_database.insert_announcement", return_value=5
    )
    stored = {
        "announcement_id": 5,
        "revisions": [],
        "sent_flag": False,
        "discord_msg_link": None,
    }
    mocker.patch(
        "moddingway.database.announcements_database.get_announcement",
        return_value=stored,
    )
    embed = discord.Embed()

    result = await announcement_service.add_announcement(
        embed, mocker.Mock(id=7), "hello", mocker.Mock()
    )

    assert result == stored
    revision = insert.call_args.kwargs["announcement_rev"]
    assert (revision.version, revision.content, revision.author_id) == (0, "hello", 7)
    assert embed_fields(embed) == [("Result", "Announcement draft created.")]


@pytest.mark.asyncio
async def test_publish_announcement_posts_latest_revision_and_marks_sent(
    mocker: MockerFixture,
):
    mocker.patch(
        "moddingway.database.announcements_database.get_announcement",
        return_value={
            "revisions": [
                {"version": 1, "content": "old", "author_id": 1},
                {"version": 2, "content": "new", "author_id": 1},
            ]
        },
    )
    set_sent = mocker.patch("moddingway.database.announcements_database.set_sent")
    channel = mocker.Mock(
        id=100, send=mocker.AsyncMock(return_value=mocker.Mock(id=200))
    )
    embed = discord.Embed()

    await announcement_service.publish_announcement(embed, channel, 5)

    assert channel.send.await_args.kwargs["embed"].description == "new"
    set_sent.assert_called_once_with(announcement_id=5, discord_msg_link="100/200")
    assert embed_fields(embed) == [("Result", "Published announcement")]


@pytest.mark.asyncio
async def test_publish_missing_announcement_does_nothing(mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.announcements_database.get_announcement", return_value=None
    )
    channel = mocker.Mock(send=mocker.AsyncMock())

    await announcement_service.publish_announcement(discord.Embed(), channel, 5)

    channel.send.assert_not_awaited()


@pytest.mark.asyncio
async def test_list_announcements_passes_status_through(mocker: MockerFixture):
    select = mocker.patch(
        "moddingway.database.announcements_database.select_announcements_bulk",
        return_value=[(1,)],
    )

    assert await announcement_service.list_announcements_service(True) == [(1,)]
    select.assert_called_once_with(True)
