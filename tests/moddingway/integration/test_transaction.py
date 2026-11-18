import asyncio

import pytest

from moddingway.database import DatabaseConnection, users_database


def discord_ids(query) -> list[str]:
    return [row[0] for row in query("SELECT discordUserID FROM users ORDER BY 1")]


async def test_transaction__rolled_back_on_error(database: DatabaseConnection, query):
    with pytest.raises(RuntimeError), database.transaction():
        users_database.add_user(1)
        raise RuntimeError("boom")

    assert discord_ids(query) == []


async def test_transaction__child_task_does_not_join(
    database: DatabaseConnection, query
):
    async def child() -> list[str]:
        # Inherits the ContextVar, but must use the shared autocommit connection
        users_database.add_user(2)
        return discord_ids(query)

    with pytest.raises(RuntimeError), database.transaction():
        users_database.add_user(1)
        seen_by_child = await asyncio.create_task(child())
        raise RuntimeError("roll back the parent")

    # The child neither saw the parent's uncommitted row nor was rolled back
    assert seen_by_child == ["2"]
    assert discord_ids(query) == ["2"]
