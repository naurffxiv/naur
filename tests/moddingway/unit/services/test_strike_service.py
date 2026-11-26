from datetime import UTC, datetime, timedelta

import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import constants
from moddingway.constants import PERMANENT_BAN_STRIKE_THRESHOLD
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
        # already past a threshold without crossing a new one: repeat the highest reached
        (8, 9, "7 day exile"),
        # crossing one threshold while already past a lower one: only the new one counts
        (4, 5, "3 day exile"),
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


@pytest.mark.parametrize(
    "severity,expected_temporary,expected_permanent",
    [
        (constants.StrikeSeverity.MINOR, 2 + constants.MINOR_INFRACTION_POINTS, 1),
        (
            constants.StrikeSeverity.MODERATE,
            2 + constants.MODERATE_INFRACTION_POINTS,
            1,
        ),
        (
            constants.StrikeSeverity.SERIOUS,
            2,
            1 + constants.SERIOUS_INFRACTION_POINTS,
        ),
    ],
)
def test_apply_strike_point_penalty(
    severity, expected_temporary, expected_permanent, create_db_user
):
    db_user = create_db_user(temporary_points=2, permanent_points=1)

    strike_service._apply_strike_point_penalty(db_user, severity)

    assert db_user.temporary_points == expected_temporary
    assert db_user.permanent_points == expected_permanent


@pytest.mark.asyncio
async def test_apply_punishment__passes_reasons(mocker: MockerFixture, create_db_user):
    user = mocker.Mock()
    embed = mocker.Mock(spec=discord.Embed)
    ban_member = mocker.patch("moddingway.services.ban_service.ban_member")
    exile_user = mocker.patch("moddingway.services.exile_service.exile_user")

    await strike_service._apply_punishment(
        embed, user, create_db_user(get_strike_points=3), 0
    )
    await strike_service._apply_punishment(
        embed,
        user,
        create_db_user(get_strike_points=PERMANENT_BAN_STRIKE_THRESHOLD),
        0,
    )

    exile_user.assert_called_once_with(
        embed,
        user,
        timedelta(days=1),
        "Your actions were severe or frequent enough for you to receive this exile",
    )
    ban_member.assert_called_once_with(
        embed,
        user,
        "Your strike were severe or frequent to be removed from NA Ultimate Raiding - FFXIV",
        False,
    )


@pytest.mark.asyncio
async def test_add_strike__records_strike_and_notifies_user(
    create_member, create_db_user, embed_fields, mocker: MockerFixture
):
    now = datetime(2025, 9, 1, 12, 0, 0)
    mocker.patch("moddingway.services.strike_service.datetime").now.return_value = now
    db_user = create_db_user(user_id=1, temporary_points=0, permanent_points=0)
    db_user.get_strike_points.side_effect = lambda: (
        db_user.temporary_points + db_user.permanent_points
    )
    mocker.patch("moddingway.database.users_database.get_user", return_value=None)
    add_user = mocker.patch(
        "moddingway.database.users_database.add_user", return_value=db_user
    )
    add_strike = mocker.patch(
        "moddingway.database.strikes_database.add_strike", return_value=55
    )
    update_points = mocker.patch(
        "moddingway.database.users_database.update_user_strike_points"
    )
    mocker.patch(
        "moddingway.services.strike_service._apply_punishment",
        return_value="1 day exile",
    )
    user = create_member(id=12345)
    author = mocker.Mock(spec=discord.Member, id=2)
    embed = discord.Embed()

    await strike_service.add_strike(
        embed, user, constants.StrikeSeverity.MODERATE, "spam", author
    )

    add_user.assert_called_once_with(12345)
    assert embed.description == "\nUser not found in database, creating new record"
    (strike,), _ = add_strike.call_args
    assert strike.model_dump() == {
        "strike_id": 55,
        "user_id": 1,
        "severity": constants.StrikeSeverity.MODERATE,
        "reason": "spam",
        "created_timestamp": now,
        "created_by": "2",
        "last_edited_timestamp": now,
        "last_edited_by": "2",
    }
    assert embed.footer.text == "Strike ID: 55"
    assert db_user.last_infraction_timestamp == now
    assert db_user.temporary_points == constants.MODERATE_INFRACTION_POINTS
    update_points.assert_called_once_with(db_user)
    assert embed_fields(embed) == [
        (
            "Strike Point Adjustment",
            "<@12345> was given a strike, bringing them to 3 points from 0 points",
        ),
        ("Punishment", "1 day exile"),
    ]
    user.create_dm.return_value.send.assert_awaited_once_with(
        content=(
            "Your actions in NA Ultimate Raiding - FFXIV resulted in a strike against "
            "your account. This may result in punishment depending on the frequency "
            "or severity of your strikes.\n**Reason:** spam"
        )
    )


@pytest.mark.asyncio
async def test_add_strike__dm_failure_names_strike_context(
    create_member, create_db_user, embed_fields, mocker: MockerFixture
):
    db_user = create_db_user(user_id=1, temporary_points=0, get_strike_points=0)
    mocker.patch("moddingway.database.users_database.get_user", return_value=db_user)
    mocker.patch("moddingway.database.strikes_database.add_strike", return_value=1)
    mocker.patch("moddingway.database.users_database.update_user_strike_points")
    mocker.patch(
        "moddingway.services.strike_service._apply_punishment", return_value="Nothing"
    )
    user = create_member(id=12345, allows_dms=False)
    embed = discord.Embed()

    await strike_service.add_strike(
        embed,
        user,
        constants.StrikeSeverity.MINOR,
        "spam",
        mocker.Mock(spec=discord.Member, id=2),
    )

    assert (
        "DM Status",
        "Failed to send DM to <@12345> for Strike, Cannot create DM",
    ) in embed_fields(embed)


@pytest.mark.asyncio
async def test_get_user_strikes__lists_by_database_user(
    mocker: MockerFixture, create_db_user
):
    db_user = create_db_user(
        user_id=7, temporary_points=1, permanent_points=0, get_strike_points=1
    )
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user", return_value=db_user
    )
    list_strikes = mocker.patch(
        "moddingway.database.strikes_database.list_strikes",
        return_value=[
            (
                99,
                constants.StrikeSeverity.MINOR,
                "test",
                "3",
                datetime(2025, 9, 1, tzinfo=UTC),
            )
        ],
    )

    res = await strike_service.get_user_strikes(mocker.Mock(id=12345))

    get_user.assert_called_once_with(12345)
    list_strikes.assert_called_once_with(7)
    assert "\n* ID: 99 | SEVERITY: 1 |" in res


@pytest.mark.asyncio
async def test_get_user_strikes__missing_user_or_strikes(
    mocker: MockerFixture, create_db_user
):
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user", return_value=None
    )
    mocker.patch("moddingway.database.strikes_database.list_strikes", return_value=[])

    assert (
        await strike_service.get_user_strikes(mocker.Mock(id=1))
        == "User not found in database"
    )

    get_user.return_value = create_db_user(user_id=7)
    assert (
        await strike_service.get_user_strikes(mocker.Mock(id=1))
        == "No strikes found for user"
    )
