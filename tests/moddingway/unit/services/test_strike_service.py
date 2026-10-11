from datetime import UTC, datetime, timedelta

import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import constants
from moddingway.constants import PERMANENT_BAN_STRIKE_THRESHOLD
from moddingway.database.strikes_database import StrikeDisplay
from moddingway.services import strike_service


@pytest.mark.parametrize(
    "previous_points,total_points,expected_punishment",
    [
        (0, 1, "Nothing"),
        (1, 8, "11 day exile"),
        (8, 10, "14 day exile"),
        (10, 11, "14 day exile"),
        (1, 9, "11 day exile"),
        (9, 12, "14 day exile"),
        (12, 13, "14 day exile"),
        (10, 15, "Permanent ban"),
        (0, 3, "1 day exile"),
        (10, 17, "Permanent ban"),
    ],
)
@pytest.mark.asyncio
async def test_apply_punishment(
    previous_points: int,
    total_points: int,
    expected_punishment: str,
    mocker: MockerFixture,
    create_db_user,
):
    mocked_user = mocker.Mock()
    mocked_db_user = create_db_user(get_strike_points=total_points)
    mocked_ban_member = mocker.patch("moddingway.services.ban_service.ban_member")
    mocked_ban_user = mocker.patch("moddingway.services.ban_service.ban_user")
    mocked_embed = mocker.Mock(spec=discord.Embed)

    spy_calculate_punishment = mocker.spy(strike_service, "_calculate_punishment")

    mocked_exile_user = mocker.patch("moddingway.services.exile_service.exile_user")

    res = await strike_service._apply_punishment(
        mocked_embed, mocked_user, mocked_db_user, previous_points
    )
    if total_points >= PERMANENT_BAN_STRIKE_THRESHOLD:
        mocked_ban_member.assert_called_once_with(
            mocked_embed,
            mocked_user,
            mocker.ANY,
            False,
        )
    else:
        mocked_ban_user.assert_not_called()
        spy_calculate_punishment.assert_called_once_with(previous_points, total_points)
        punishment_days = spy_calculate_punishment.spy_return
        if punishment_days == 0:
            mocked_exile_user.assert_not_called()
        else:
            mocked_exile_user.assert_called_once_with(
                mocked_embed,
                mocked_user,
                timedelta(days=punishment_days),
                mocker.ANY,
            )
    assert res == expected_punishment


@pytest.mark.parametrize(
    "user_id,total_points,strike_params,expected_result",
    [
        (
            "1",
            "1",
            [
                (
                    "1",
                    constants.StrikeSeverity.MINOR,
                    "test",
                    "1",
                    datetime(2025, 9, 1, tzinfo=UTC),
                )
            ],
            "Strikes found for <@1>: [Temporary points: None | Permanent points: None]"
            "\n* ID: 1 | SEVERITY: 1 | Moderator: <@1> | DATE ISSUED: <t:1756684800:F> | REASON: test"
            "\nTotal Points: 1",
        ),
        (
            "1",
            "3",
            [
                (
                    "1",
                    constants.StrikeSeverity.MINOR,
                    "test",
                    "1",
                    datetime(2025, 9, 1, tzinfo=UTC),
                ),
                (
                    "2",
                    constants.StrikeSeverity.MODERATE,
                    "test2",
                    "3",
                    datetime(2025, 9, 2, tzinfo=UTC),
                ),
            ],
            "Strikes found for <@1>: [Temporary points: None | Permanent points: None]"
            "\n* ID: 1 | SEVERITY: 1 | Moderator: <@1> | DATE ISSUED: <t:1756684800:F> | REASON: test"
            "\n* ID: 2 | SEVERITY: 2 | Moderator: <@3> | DATE ISSUED: <t:1756771200:F> | REASON: test2"
            "\nTotal Points: 3",
        ),
    ],
)
@pytest.mark.asyncio
async def test_get_user_strikes(
    user_id: str,
    total_points: str,
    strike_params: list[tuple],
    expected_result: str,
    mocker: MockerFixture,
    create_db_user,
):
    mocked_user = mocker.Mock(id=user_id)
    mocked_db_user = create_db_user(get_strike_points=total_points)
    mocker.patch(
        "moddingway.database.users_database.get_user", return_value=mocked_db_user
    )
    mocker.patch(
        "moddingway.database.strikes_database.list_strikes", return_value=strike_params
    )
    res = await strike_service.get_user_strikes(mocked_user)
    assert res == expected_result


@pytest.mark.asyncio
async def test_add_strike(
    create_db_user, create_member, create_embed, mocker: MockerFixture
):
    mocked_db_user = create_db_user(user_id=1, temporary_points=0, get_strike_points=0)

    mocked_user = create_member(id=1)

    mocked_author = mocker.Mock(spec=discord.Member)
    mocked_author.id = 2

    mocked_logging_embed = create_embed()
    mocked_logging_embed.add_field = mocker.Mock()
    mocked_logging_embed.set_footer = mocker.Mock()

    mocked_get_user = mocker.patch(
        "moddingway.database.users_database.get_user", return_value=mocked_db_user
    )
    mocked_add_strike = mocker.patch(
        "moddingway.database.strikes_database.add_strike", return_value=1
    )

    mocked_strike = mocker.Mock()
    mocked_strike_class = mocker.patch(
        "moddingway.services.strike_service.Strike",
        autospec=True,
        return_value=mocked_strike,
    )

    mocked_update_user_strike_points = mocker.patch(
        "moddingway.database.users_database.update_user_strike_points",
        return_value=None,
    )

    mocked_punishment = mocker.Mock()
    mocked__apply_punishment = mocker.patch(
        "moddingway.services.strike_service._apply_punishment",
        return_value=mocked_punishment,
    )

    await strike_service.add_strike(
        logging_embed=mocked_logging_embed,
        user=mocked_user,
        severity=constants.StrikeSeverity.MODERATE,
        reason="test",
        author=mocked_author,
    )

    mocked_get_user.assert_called_with(mocked_user.id)

    mocked_strike_class.assert_called_with(
        user_id=mocked_user.id,
        severity=constants.StrikeSeverity.MODERATE,
        reason="test",
        created_timestamp=mocker.ANY,
        created_by=str(mocked_author.id),
        last_edited_timestamp=mocker.ANY,
        last_edited_by=str(mocked_author.id),
    )

    mocked_add_strike.assert_called_with(mocked_strike)

    mocked_update_user_strike_points.assert_called_with(mocked_db_user)

    mocked__apply_punishment.assert_called_with(
        mocked_logging_embed, mocked_user, mocked_db_user, 0
    )


@pytest.mark.asyncio
async def test_edit_strike__not_found(create_embed, mocker: MockerFixture):
    mocked_logging_embed = create_embed()
    mocker.patch("moddingway.database.strikes_database.get_strike", return_value=None)

    result = await strike_service.edit_strike(
        mocked_logging_embed,
        mocker.Mock(id=99),
        mocker.Mock(),
        1,
        "updated reason",
        None,
    )

    assert result == "Strike not found in database"


@pytest.mark.asyncio
async def test_edit_strike__no_updates(create_embed, mocker: MockerFixture):
    mocked_logging_embed = create_embed()
    mocker.patch(
        "moddingway.database.strikes_database.get_strike",
        return_value=StrikeDisplay(
            strike_id=1,
            user_id=10,
            discord_user_id="123",
            severity=constants.StrikeSeverity.MINOR,
            reason="original",
            created_by="1",
            created_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
            last_edited_by="1",
            last_edited_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
        ),
    )

    result = await strike_service.edit_strike(
        mocked_logging_embed,
        mocker.Mock(id=99),
        mocker.Mock(),
        1,
        None,
        None,
    )

    assert result == "No updates provided. Please specify a reason and/or severity."


@pytest.mark.asyncio
async def test_edit_strike__reason_too_long(create_embed, mocker: MockerFixture):
    mocked_logging_embed = create_embed()
    mocker.patch(
        "moddingway.database.strikes_database.get_strike",
        return_value=StrikeDisplay(
            strike_id=1,
            user_id=10,
            discord_user_id="123",
            severity=constants.StrikeSeverity.MINOR,
            reason="original",
            created_by="1",
            created_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
            last_edited_by="1",
            last_edited_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
        ),
    )

    result = await strike_service.edit_strike(
        mocked_logging_embed,
        mocker.Mock(id=99),
        mocker.Mock(),
        1,
        "x" * 301,
        None,
    )

    assert "Reason is too long" in result


@pytest.mark.asyncio
async def test_edit_strike__updates_reason_and_dms_user(
    create_embed, mocker: MockerFixture
):
    mocked_logging_embed = create_embed()
    mocked_logging_embed.add_field = mocker.Mock()
    mocked_logging_embed.set_footer = mocker.Mock()

    mocker.patch(
        "moddingway.database.strikes_database.get_strike",
        return_value=StrikeDisplay(
            strike_id=1,
            user_id=10,
            discord_user_id="123",
            severity=constants.StrikeSeverity.MINOR,
            reason="original",
            created_by="1",
            created_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
            last_edited_by="1",
            last_edited_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
        ),
    )
    mocked_update_strike = mocker.patch(
        "moddingway.database.strikes_database.update_strike", return_value=True
    )
    mocked_fetch_user = mocker.AsyncMock()
    mocked_client = mocker.Mock(fetch_user=mocked_fetch_user)
    mocked_send_dm = mocker.patch("moddingway.services.strike_service.send_dm")

    editor = mocker.Mock(id=99)

    result = await strike_service.edit_strike(
        mocked_logging_embed,
        editor,
        mocked_client,
        1,
        "updated reason",
        None,
    )

    assert result == "Strike successfully updated"
    mocked_update_strike.assert_called_once_with(
        strike_id=1,
        reason="updated reason",
        severity=constants.StrikeSeverity.MINOR,
        editor_id="99",
        update_timestamp=mocker.ANY,
    )
    mocked_fetch_user.assert_called_once_with(123)
    mocked_send_dm.assert_called_once()


@pytest.mark.asyncio
async def test_edit_strike__updates_severity_and_points(
    create_embed, create_db_user, mocker: MockerFixture
):
    mocked_logging_embed = create_embed()
    mocked_logging_embed.add_field = mocker.Mock()
    mocked_logging_embed.set_footer = mocker.Mock()

    mocker.patch(
        "moddingway.database.strikes_database.get_strike",
        return_value=StrikeDisplay(
            strike_id=1,
            user_id=10,
            discord_user_id="123",
            severity=constants.StrikeSeverity.MINOR,
            reason="original",
            created_by="1",
            created_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
            last_edited_by="1",
            last_edited_timestamp=datetime(2025, 9, 1, tzinfo=UTC),
        ),
    )
    db_user = create_db_user(
        user_id=10, temporary_points=2, permanent_points=0, get_strike_points=5
    )
    mocker.patch(
        "moddingway.database.users_database.get_user_by_internal_id",
        return_value=db_user,
    )
    mocked_decrement = mocker.patch(
        "moddingway.database.users_database.decrement_user_strike_points"
    )
    mocked_update_points = mocker.patch(
        "moddingway.database.users_database.update_user_strike_points"
    )
    mocker.patch(
        "moddingway.database.strikes_database.update_strike", return_value=True
    )
    mocked_client = mocker.Mock(fetch_user=mocker.AsyncMock())
    mocker.patch("moddingway.services.strike_service.send_dm")

    result = await strike_service.edit_strike(
        mocked_logging_embed,
        mocker.Mock(id=99),
        mocked_client,
        1,
        None,
        constants.StrikeSeverity.MODERATE,
    )

    assert result == "Strike successfully updated"
    mocked_decrement.assert_called_once_with(10, 1, 0)
    mocked_update_points.assert_called_once_with(db_user)
    assert db_user.temporary_points == 5
