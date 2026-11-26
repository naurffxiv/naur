import discord
import pytest
from command_helpers import reply, run, sent, target
from pytest_mock.plugin import MockerFixture

from moddingway.commands import announcement_commands
from moddingway.commands.announcement_commands import (
    AnnouncementEditModal,
    AnnouncementPaginator,
    AnnouncementPublishView,
    AnnouncementShowView,
    create_announcement_commands,
)
from moddingway.constants import (
    MAX_CHAR_LIMIT,
    MESSAGE_LINK_PREFIX,
    TRUNCATE_OFFSET,
    Role,
)

GUILD = announcement_commands.settings.guild_id


def announcement(announcement_id=5, revisions=None, sent_flag=False, link=None):
    return {
        "announcement_id": announcement_id,
        "revisions": revisions or [{"version": 1, "content": "hello", "author_id": 7}],
        "sent_flag": sent_flag,
        "discord_msg_link": link,
    }


@pytest.fixture
def ann_bot(bot):
    create_announcement_commands(bot)
    return bot


@pytest.mark.asyncio
async def test_draft_announcement_posts_the_draft_for_review(
    ann_bot, make_interaction, mocker: MockerFixture, monkeypatch
):
    mocker.patch(
        "moddingway.services.announcement_service.add_announcement",
        return_value=announcement(),
    )
    draft_channel = mocker.Mock(spec=discord.TextChannel, send=mocker.AsyncMock())
    monkeypatch.setattr(ann_bot, "get_channel", lambda _id: draft_channel)
    i = make_interaction("draft_announcement")

    await run(ann_bot, "draft_announcement", i, announcement_text="hello")

    embed = draft_channel.send.await_args.kwargs["embed"]
    assert embed.title == "Announcement Draft (Rev 1/1)"
    assert embed.description == "hello"
    assert isinstance(
        draft_channel.send.await_args.kwargs["view"], AnnouncementShowView
    )
    assert reply(i) == "Successfully added announcement. (ID: 5)"


@pytest.mark.asyncio
async def test_draft_announcement_fails_without_a_draft_channel(
    ann_bot, make_interaction, mocker: MockerFixture, monkeypatch
):
    mocker.patch(
        "moddingway.services.announcement_service.add_announcement",
        return_value=announcement(),
    )
    monkeypatch.setattr(ann_bot, "get_channel", lambda _id: None)
    i = make_interaction("draft_announcement")

    with pytest.raises(ValueError, match="Could not find announcement channel"):
        await run(ann_bot, "draft_announcement", i, announcement_text="hello")
    assert reply(i) == "\nAn unexpected error has occurred"


@pytest.mark.parametrize(
    "stored,expected,has_view",
    [
        (None, "Announcement not found.", False),
        (
            announcement(sent_flag=True, link="10/20"),
            f"Announcement already published: https://discord.com/channels/{GUILD}/10/20",
            False,
        ),
        (
            announcement(),
            "Are you sure you want to publish this announcement (ID: 5)",
            True,
        ),
    ],
)
@pytest.mark.asyncio
async def test_publish_announcement(
    ann_bot, make_interaction, mocker: MockerFixture, stored, expected, has_view
):
    mocker.patch(
        "moddingway.database.announcements_database.get_announcement",
        return_value=stored,
    )
    i = make_interaction("publish_announcement")

    await run(
        ann_bot,
        "publish_announcement",
        i,
        channel=mocker.Mock(spec=discord.TextChannel),
        announcement_id=5,
    )

    assert sent(i) == expected
    assert (
        isinstance(
            i.response.send_message.await_args.kwargs.get("view"),
            AnnouncementPublishView,
        )
        is has_view
    )


@pytest.mark.asyncio
async def test_publish_view_confirm_and_cancel(make_interaction, mocker: MockerFixture):
    publish = mocker.patch(
        "moddingway.services.announcement_service.publish_announcement"
    )
    channel = mocker.Mock(spec=discord.TextChannel, mention="#news")
    original = make_interaction("publish_announcement")

    confirm = make_interaction()
    await AnnouncementPublishView(
        5, channel, original, None
    ).confirm_button_callback.callback(confirm)
    assert publish.await_args.kwargs == {"channel": channel, "announcement_id": 5}
    assert reply(confirm) == "Successfully published announcement. (ID: 5)"

    cancel = make_interaction()
    await AnnouncementPublishView(
        5, channel, original, None
    ).cancel_button_callback.callback(cancel)
    assert reply(cancel) == "Announcement publishing cancelled."


@pytest.mark.asyncio
async def test_list_announcements(ann_bot, make_interaction, mocker: MockerFixture):
    listing = mocker.patch(
        "moddingway.services.announcement_service.list_announcements_service",
        return_value=[],
    )
    i = make_interaction("list_announcements")
    await run(ann_bot, "list_announcements", i, ephemeral=True, sent_status=None)
    assert sent(i) == "Can't find announcements with selected parameters."

    listing.return_value = [
        (1, [{"version": 1, "content": "a", "author_id": 7}], False, None)
    ]
    i = make_interaction("list_announcements")
    await run(ann_bot, "list_announcements", i, ephemeral=False, sent_status=False)
    kwargs = i.response.send_message.await_args.kwargs
    assert kwargs["embed"].title == "Unsent Announcements"
    assert kwargs["ephemeral"] is False
    assert kwargs["view"].message is i.original


def _rows(n):
    return [
        (
            k,
            [{"version": 1, "content": f"announcement {k}", "author_id": 7}],
            k % 2 == 0,
            f"10/{k}",
        )
        for k in range(1, n + 1)
    ]


@pytest.mark.asyncio
async def test_paginator_pages_six_at_a_time(make_interaction, mocker: MockerFixture):
    author = mocker.Mock()
    view = AnnouncementPaginator(_rows(8), None, author)

    first = view.create_embed()
    assert first.title == "All Announcements"
    assert first.footer.text == "Page 1/2"
    assert len(first.fields) == 6 * 4
    assert first.fields[2].value == "Unsent"
    assert first.fields[6].value == f"Sent: {MESSAGE_LINK_PREFIX}{GUILD}/10/2"
    assert view.prev_button.disabled and not view.next_button.disabled

    click = make_interaction(user=author)
    await view.next_button.callback(click)
    assert view.current_page == view.total_pages
    assert (
        click.response.edit_message.await_args.kwargs["embed"].footer.text == "Page 2/2"
    )
    assert view.next_button.disabled and not view.prev_button.disabled

    stranger = make_interaction()
    await view.prev_button.callback(stranger)
    assert sent(stranger) == "This isn't your menu!"
    assert view.current_page == view.total_pages


@pytest.mark.asyncio
async def test_paginator_truncates_long_announcements(mocker: MockerFixture):
    long_text = "x" * (MAX_CHAR_LIMIT + 50)
    view = AnnouncementPaginator(
        [(1, [{"version": 1, "content": long_text, "author_id": 7}], True, "1/2")],
        True,
        mocker.Mock(),
    )

    preview = view.create_embed().fields[3].value

    assert view.create_embed().title == "Sent Announcements"
    assert preview == "x" * (MAX_CHAR_LIMIT - TRUNCATE_OFFSET) + "..."


@pytest.mark.asyncio
async def test_show_announcement(ann_bot, make_interaction, mocker: MockerFixture):
    get = mocker.patch(
        "moddingway.database.announcements_database.get_announcement", return_value=None
    )
    i = make_interaction("show_announcement")
    await run(ann_bot, "show_announcement", i, announcement_id=5)
    assert sent(i) == "Announcement not found."

    get.return_value = announcement()
    i = make_interaction("show_announcement")
    await run(ann_bot, "show_announcement", i, announcement_id=5)
    assert (
        i.response.send_message.await_args.kwargs["embed"].footer.text
        == "Announcement ID 5"
    )


def _revisions(n):
    return [
        {"version": k, "content": f"rev {k}", "author_id": k} for k in range(1, n + 1)
    ]


@pytest.mark.asyncio
async def test_show_view_steps_through_revisions(
    make_interaction, mocker: MockerFixture
):
    author = mocker.Mock()
    view = AnnouncementShowView(announcement(revisions=_revisions(3)), author, None)
    assert view.create_embed().description == "rev 3"
    assert view.right_button.disabled and not view.left_button.disabled

    click = make_interaction(user=author)
    await view.left_button.callback(click)
    embed = click.response.edit_message.await_args.kwargs["embed"]
    assert (embed.title, embed.description) == ("Announcement Draft (Rev 2/3)", "rev 2")
    assert embed.fields[0].value == "<@2>"

    await view.right_button.callback(make_interaction(user=author))
    assert view.current_rev_index == len(view.revisions) - 1

    stranger = make_interaction()
    await view.left_button.callback(stranger)
    assert sent(stranger) == "This isn't your menu!"


@pytest.mark.asyncio
async def test_edit_button_is_admin_only_and_blocked_after_sending(
    make_interaction, mocker: MockerFixture
):
    view = AnnouncementShowView(announcement(), mocker.Mock(), None)

    mod = make_interaction(user=target(mocker, role=Role.MOD))
    await view.edit_button.callback(mod)
    assert sent(mod) == "Only admins are allowed to edit announcements."

    admin = make_interaction(user=target(mocker, role=Role.ADMIN))
    await view.edit_button.callback(admin)
    modal = admin.response.send_modal.await_args.args[0]
    assert isinstance(modal, AnnouncementEditModal)
    assert modal.announcement_content.default == "hello"

    sent_view = AnnouncementShowView(
        announcement(sent_flag=True, link="1/2"), mocker.Mock(), None
    )
    assert sent_view.edit_button.disabled
    admin = make_interaction(user=target(mocker, role=Role.ADMIN))
    await sent_view.edit_button.callback(admin)
    assert (
        sent(admin) == "This announcement has already been sent and cannot be edited."
    )


@pytest.mark.asyncio
async def test_edit_modal_saves_a_new_revision(make_interaction, mocker: MockerFixture):
    add_revision = mocker.patch(
        "moddingway.database.announcements_database.add_revision"
    )
    mocker.patch(
        "moddingway.database.announcements_database.get_announcement",
        return_value=announcement(revisions=_revisions(2)),
    )
    modal = AnnouncementEditModal(announcement_id=5, current_content="old", bot=None)
    i = make_interaction(user=mocker.Mock(id=9))

    modal.announcement_content._value = "new text"  # what the admin typed
    await modal.on_submit(i)

    add_revision.assert_called_once_with(
        announcement_id=5, author_id=9, content="new text"
    )
    assert i.response.edit_message.await_args.kwargs["embed"].description == "rev 2"
    assert (
        i.followup.send.await_args.args[0]
        == "Successfully updated announcement (ID: 5)."
    )

    modal.announcement_content._value = "   "
    blank = make_interaction()
    await modal.on_submit(blank)
    assert sent(blank) == "Announcement content cannot be empty."
    add_revision.assert_called_once()
