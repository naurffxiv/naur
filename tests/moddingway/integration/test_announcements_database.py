import pytest

from moddingway.constants import ANNOUNCEMENT_REV_LIMIT
from moddingway.database import announcements_database
from moddingway.database.models import AnnouncementRevision


def test_insert_and_get_announcement():
    announcement_id = announcements_database.insert_announcement(
        AnnouncementRevision(version=1, content="hello", author_id=7)
    )

    assert announcements_database.get_announcement(announcement_id) == {
        "announcement_id": announcement_id,
        "revisions": [{"version": 1, "content": "hello", "author_id": 7}],
        "sent_flag": False,
        "discord_msg_link": None,
    }
    assert announcements_database.get_announcement(999) is None


def test_add_revision_keeps_only_the_latest_revisions():
    announcement_id = announcements_database.insert_announcement(
        AnnouncementRevision(version=1, content="v1", author_id=7)
    )

    for n in range(2, ANNOUNCEMENT_REV_LIMIT + 3):
        assert announcements_database.add_revision(announcement_id, 8, f"v{n}") is True

    revisions = announcements_database.get_announcement(announcement_id)["revisions"]
    assert len(revisions) == ANNOUNCEMENT_REV_LIMIT
    assert revisions[-1] == {
        "version": ANNOUNCEMENT_REV_LIMIT + 2,
        "author_id": 8,
        "content": f"v{ANNOUNCEMENT_REV_LIMIT + 2}",
    }
    assert [r["version"] for r in revisions] == list(
        range(3, ANNOUNCEMENT_REV_LIMIT + 3)
    )

    with pytest.raises(ValueError, match=r"Announcement ID 999 not found\."):
        announcements_database.add_revision(999, 8, "x")


def test_set_sent_and_bulk_select_by_status():
    sent = announcements_database.insert_announcement(
        AnnouncementRevision(version=1, content="a", author_id=1)
    )
    draft = announcements_database.insert_announcement(
        AnnouncementRevision(version=1, content="b", author_id=1)
    )

    assert (
        announcements_database.set_sent(sent, "https://discord.com/channels/1/2/3")
        is True
    )

    assert [
        row[0] for row in announcements_database.select_announcements_bulk(True)
    ] == [sent]
    assert [
        row[0] for row in announcements_database.select_announcements_bulk(False)
    ] == [draft]
    assert sorted(
        row[0] for row in announcements_database.select_announcements_bulk()
    ) == [sent, draft]
    assert (
        announcements_database.get_announcement(sent)["discord_msg_link"]
        == "https://discord.com/channels/1/2/3"
    )
    with pytest.raises(ValueError, match="Announcement ID 999 not found"):
        announcements_database.set_sent(999, "x")
