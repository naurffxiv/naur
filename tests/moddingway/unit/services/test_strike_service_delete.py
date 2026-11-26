import discord
import pytest
from pytest_mock.plugin import MockerFixture

from moddingway import constants
from moddingway.services import strike_service


@pytest.mark.parametrize(
    "severity,temporary,permanent",
    [
        (constants.StrikeSeverity.MINOR, constants.MINOR_INFRACTION_POINTS, 0),
        (constants.StrikeSeverity.MODERATE, constants.MODERATE_INFRACTION_POINTS, 0),
        (constants.StrikeSeverity.SERIOUS, 0, constants.SERIOUS_INFRACTION_POINTS),
    ],
)
@pytest.mark.asyncio
async def test_delete_strike_refunds_its_points(
    mocker: MockerFixture, severity, temporary, permanent
):
    mocker.patch(
        "moddingway.database.strikes_database.delete_strike",
        return_value=(5, "1", str(severity.value)),
    )
    decrement = mocker.patch(
        "moddingway.database.users_database.decrement_user_strike_points"
    )

    assert (
        await strike_service.delete_strike(discord.Embed(), 5)
        == "Successfully deleted strike 5"
    )

    decrement.assert_called_once_with(1, temporary, permanent)


@pytest.mark.asyncio
async def test_delete_missing_strike(mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.strikes_database.delete_strike", return_value=None
    )
    decrement = mocker.patch(
        "moddingway.database.users_database.decrement_user_strike_points"
    )

    assert (
        await strike_service.delete_strike(discord.Embed(), 5)
        == "Strike not found, no change will be made"
    )
    decrement.assert_not_called()
