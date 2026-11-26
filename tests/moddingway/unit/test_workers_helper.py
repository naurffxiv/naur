from datetime import UTC, datetime, timedelta

import discord
import pytest
from discord.utils import time_snowflake
from pytest_mock.plugin import MockerFixture

from moddingway.workers import helper


def _id(age: timedelta) -> int:
    return time_snowflake(datetime.now(UTC) - age)


class _Messages:
    def __init__(self, messages):
        self._messages = messages

    def __aiter__(self):
        self._it = iter(self._messages)
        return self

    async def __anext__(self):
        try:
            return next(self._it)
        except StopIteration:
            raise StopAsyncIteration from None


@pytest.fixture(autouse=True)
def no_sleep(mocker: MockerFixture):
    mocker.patch("moddingway.workers.helper.asyncio.sleep", new=mocker.AsyncMock())


@pytest.mark.asyncio
async def test_automod_channel_deletes_old_unpinned_messages(mocker: MockerFixture):
    old = mocker.Mock(
        id=_id(timedelta(minutes=200)), pinned=False, delete=mocker.AsyncMock()
    )
    pinned = mocker.Mock(
        id=_id(timedelta(minutes=200)), pinned=True, delete=mocker.AsyncMock()
    )
    fresh = mocker.Mock(
        id=_id(timedelta(minutes=10)), pinned=False, delete=mocker.AsyncMock()
    )
    failing = mocker.Mock(
        id=_id(timedelta(minutes=200)),
        pinned=False,
        delete=mocker.AsyncMock(side_effect=RuntimeError("gone")),
    )

    removed, errors = await helper.automod_channel(
        _Messages([old, pinned, fresh, failing]), 150, "pf-recruitment"
    )

    assert (removed, errors) == (1, 1)
    old.delete.assert_awaited_once()
    pinned.delete.assert_not_awaited()
    fresh.delete.assert_not_awaited()


@pytest.mark.parametrize(
    "pinned,age,delete_error,expected",
    [
        (False, timedelta(days=8), None, (1, 0)),
        (False, timedelta(days=6), None, (0, 0)),
        (True, timedelta(days=30), None, (0, 0)),
        (False, timedelta(days=8), RuntimeError("no perms"), (0, 1)),
    ],
)
@pytest.mark.asyncio
async def test_automod_thread(
    mocker: MockerFixture, pinned, age, delete_error, expected
):
    thread = mocker.Mock(
        flags=mocker.Mock(pinned=pinned),
        last_message_id=_id(age),
        delete=mocker.AsyncMock(side_effect=delete_error),
    )

    assert await helper.automod_thread(thread, 7, 0, 0) == expected


@pytest.mark.asyncio
async def test_automod_thread_without_messages_is_kept(mocker: MockerFixture):
    thread = mocker.Mock(
        flags=mocker.Mock(pinned=False), last_message_id=None, delete=mocker.AsyncMock()
    )

    assert await helper.automod_thread(thread, 7, 2, 1) == (2, 1)
    thread.delete.assert_not_awaited()


@pytest.mark.asyncio
async def test_automod_embeds_describe_the_cleanup(mocker: MockerFixture):
    bot = mocker.Mock(
        user=mocker.Mock(
            spec=discord.ClientUser,
            display_name="bot",
            display_avatar=mocker.Mock(url="u"),
        )
    )
    create = mocker.patch("moddingway.workers.helper.create_interaction_embed_context")
    when = datetime(2026, 1, 1, tzinfo=UTC)

    helper.create_automod_embed(bot, 10, 3, 1, when)
    helper.create_channel_automod_embed(bot, 10, 4, 0, when)
    helper.create_autounexile_embed(bot, None, 12345, "9", when)

    descriptions = [c.kwargs["description"] for c in create.call_args_list]
    assert descriptions == [
        "Successfully removed 3 inactive thread(s) from <#10>.\n1 inactive thread(s) failed to be removed.",
        "Successfully removed 4 old message(s) from <#10>.\n0 old messages(s) failed to be removed.",
        "<@12345>'s exile has timed out",
    ]
    assert create.call_args_list[2].kwargs["footer"] == "Exile ID: 9"
