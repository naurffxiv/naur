import discord
import pytest

from moddingway.services import exile_service, exiles
from moddingway.services.exiles import ExileOutcome, ExileResult


@pytest.mark.parametrize(
    "duration,expected",
    [
        ("1sec", "1 second"),
        ("30min", "30 minutes"),
        ("1hour", "1 hour"),
        ("7day", "7 days"),
        ("bad", None),
    ],
)
def test_format_time_string(duration, expected):
    assert exile_service.format_time_string(duration) == expected


@pytest.mark.parametrize("outcome", [ExileOutcome.FAILED])
def test_log_exile_result__release_outcomes_have_a_message(outcome):
    embed = discord.Embed()

    exile_service.log_exile_result(
        embed, ExileResult(outcome, discord_id=1, exile_id=7, error="boom")
    )

    assert embed.fields
    assert embed.footer.text == "Exile ID: 7"


async def test_unexile__member_without_exiled_role_is_refused(create_member):
    result = await exiles.unexile(create_member())

    assert result.outcome == ExileOutcome.NOT_EXILED
