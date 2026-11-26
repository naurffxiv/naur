import pytest


@pytest.fixture(autouse=True)
def mock_datetime_now():
    """Event handlers build embeds with real timestamps, which discord.py type-checks."""
