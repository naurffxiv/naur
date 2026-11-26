from fastapi.testclient import TestClient

from moddingway.database import users_database
from moddingway_api import main as api_main


def test_users_endpoint_pages_through_the_database():
    for discord_id in range(1, 6):
        users_database.add_user(discord_id)
    client = TestClient(api_main.app)

    pages = [
        client.get("/users", params={"page": n, "size": 2}).json() for n in (1, 2, 3)
    ]

    assert [p["total"] for p in pages] == [5, 5, 5]
    assert [[u["userID"] for u in p["items"]] for p in pages] == [
        ["1", "2"],
        ["3", "4"],
        ["5"],
    ]
