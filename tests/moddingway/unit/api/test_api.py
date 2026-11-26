from http import HTTPStatus

import httpx
import pytest
from fakes import make_user
from fastapi.testclient import TestClient
from pytest_mock.plugin import MockerFixture

from moddingway.constants import MAX_APPEAL_REASON_LENGTH, UserRole
from moddingway.database.models import BanForm
from moddingway_api import main as api_main
from moddingway_api.routes import banneduser_routes

DISCORD_ID = "123456789012345678"


@pytest.fixture
def client():
    # No `with`: the lifespan (which connects to Postgres) does not run.
    return TestClient(api_main.app)


def test_root_redirects_to_docs_and_health_is_ok(client):
    assert client.get("/", follow_redirects=False).headers["location"] == "/docs"
    assert client.get("/health").json() == {"status": "ok"}


def test_get_user(client, mocker: MockerFixture):
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=make_user(user_id=7, role=UserRole.MOD, temporary=2, permanent=7),
    )

    assert client.get("/users/555").json() == {
        "userID": "7",
        "isMod": True,
        "strikePoints": 9,
    }
    get_user.assert_called_once_with(555)

    get_user.return_value = None
    response = client.get("/users/555")
    assert (response.status_code, response.json()) == (
        404,
        {"detail": "User not found"},
    )


def test_list_users_is_paginated(client, mocker: MockerFixture):
    get_users = mocker.patch(
        "moddingway.database.users_database.get_users",
        return_value=[make_user(user_id=3), make_user(user_id=4)],
    )
    mocker.patch("moddingway.database.users_database.get_user_count", return_value=5)

    body = client.get("/users", params={"page": 2, "size": 2}).json()

    get_users.assert_called_once_with(2, 2)
    assert (body["total"], body["page"], body["pages"]) == (5, 2, 3)
    assert [u["userID"] for u in body["items"]] == ["3", "4"]


def test_get_mod(client, mocker: MockerFixture):
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=make_user(user_id=7, role=UserRole.SYSADMIN),
    )
    assert client.get("/mods/555").json() == {"modID": "7"}

    get_user.return_value = make_user(user_id=7, role=UserRole.USER)
    assert client.get("/mods/555").json() == {"detail": "Mod not found"}

    get_user.return_value = None
    assert client.get("/mods/555").status_code == HTTPStatus.NOT_FOUND


def test_list_mods_and_banned_users(client, mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.users_database.get_mods",
        return_value=[make_user(user_id=2, role=UserRole.MOD)],
    )
    mocker.patch("moddingway.database.users_database.get_mod_count", return_value=1)
    mocker.patch(
        "moddingway.database.users_database.get_banned_users",
        return_value=[make_user(user_id=3, banned=True)],
    )
    mocker.patch("moddingway.database.users_database.get_banned_count", return_value=1)

    assert client.get("/mods").json()["items"] == [{"modID": "2"}]
    assert client.get("/bannedUsers").json()["items"] == [{"userID": "3"}]


@pytest.fixture
def discord_api(monkeypatch):
    """Route the routes' httpx calls to an in-memory Discord."""
    calls = []
    status = {"code": 204}

    def handler(request: httpx.Request):
        calls.append(request)
        return httpx.Response(status["code"])

    real_client = httpx.AsyncClient
    monkeypatch.setattr(
        banneduser_routes.httpx,
        "AsyncClient",
        lambda: real_client(transport=httpx.MockTransport(handler)),
    )
    return calls, status


def test_ban_user_calls_discord(client, discord_api):
    calls, _ = discord_api

    response = client.post(
        "/bannedUsers",
        json={"user_id": DISCORD_ID, "reason": "spam", "delete_message_days": 1},
    )

    assert response.json() == {"detail": f"User {DISCORD_ID} has been banned"}
    request = calls[0]
    assert request.method == "PUT"
    assert (
        request.url.path
        == f"/api/v10/guilds/{banneduser_routes.settings.guild_id}/bans/{DISCORD_ID}"
    )
    assert request.headers["Authorization"].startswith("Bot ")
    assert request.read() == b'{"reason":"spam","delete_message_days":1}'


def test_unban_user_calls_discord(client, discord_api):
    calls, _ = discord_api

    assert client.delete("/bannedUsers", params={"user_id": DISCORD_ID}).json() == {
        "detail": f"User {DISCORD_ID} has been unbanned"
    }
    assert calls[0].method == "DELETE"


@pytest.mark.parametrize("bad_id", ["abc", "123", "1" * 22])
def test_ban_endpoints_validate_discord_ids(client, discord_api, bad_id):
    calls, _ = discord_api

    assert (
        client.post("/bannedUsers", json={"user_id": bad_id}).status_code
        == HTTPStatus.BAD_REQUEST
    )
    assert (
        client.delete("/bannedUsers", params={"user_id": bad_id}).status_code
        == HTTPStatus.BAD_REQUEST
    )
    assert calls == []


def test_ban_endpoints_report_discord_failures(client, discord_api):
    _, status = discord_api
    status["code"] = 403

    assert (
        client.post("/bannedUsers", json={"user_id": DISCORD_ID}).status_code
        == HTTPStatus.SERVICE_UNAVAILABLE
    )
    assert (
        client.delete("/bannedUsers", params={"user_id": DISCORD_ID}).status_code
        == HTTPStatus.SERVICE_UNAVAILABLE
    )


FORM = BanForm(
    form_id=4, user_id=1, reason="sorry", submission_timestamp="2026-01-01T00:00:00"
)


def test_get_ban_forms(client, mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.banforms_database.get_ban_forms", return_value=[FORM]
    )
    mocker.patch("moddingway.database.banforms_database.get_form_count", return_value=1)
    get_form = mocker.patch(
        "moddingway.database.banforms_database.get_ban_form", return_value=FORM
    )

    assert client.get("/banforms").json()["items"][0]["reason"] == "sorry"
    assert client.get("/banforms/4").json()["form_id"] == FORM.form_id
    get_form.return_value = None
    assert client.get("/banforms/4").status_code == HTTPStatus.NOT_FOUND


def test_submit_form(client, mocker: MockerFixture):
    get_user = mocker.patch(
        "moddingway.database.users_database.get_user",
        return_value=make_user(user_id=1, banned=True),
    )
    add_form = mocker.patch(
        "moddingway.database.banforms_database.add_form", return_value=4
    )

    assert client.post(
        "/banforms", json={"user_id": DISCORD_ID, "reason": "sorry"}
    ).json() == {"detail": f"User {DISCORD_ID}'s form has been submitted"}
    get_user.assert_called_once_with(int(DISCORD_ID))
    submitted = add_form.call_args.args[0]
    assert (submitted.user_id, submitted.reason) == (1, "sorry")

    too_long = client.post(
        "/banforms",
        json={"user_id": DISCORD_ID, "reason": "x" * (MAX_APPEAL_REASON_LENGTH + 1)},
    )
    assert too_long.json()["detail"].startswith("Appeal reason max length is")

    get_user.return_value = make_user(user_id=1, banned=False)
    assert client.post(
        "/banforms", json={"user_id": DISCORD_ID, "reason": "sorry"}
    ).json() == {"detail": f"User {DISCORD_ID} is not currently banned"}
    assert add_form.call_count == 1


def test_submit_form_for_unknown_user_is_not_found(client, mocker: MockerFixture):
    mocker.patch("moddingway.database.users_database.get_user", return_value=None)
    add_form = mocker.patch("moddingway.database.banforms_database.add_form")

    response = client.post("/banforms", json={"user_id": DISCORD_ID, "reason": "sorry"})

    assert (response.status_code, response.json()) == (
        404,
        {"detail": "User not found"},
    )
    add_form.assert_not_called()


def test_approving_a_form_unbans_the_user(client, mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.banforms_database.get_ban_form", return_value=FORM
    )
    update = mocker.patch(
        "moddingway.database.banforms_database.update_form", return_value=(True, 4)
    )
    mocker.patch(
        "moddingway.database.banforms_database.get_user_from_form",
        return_value=DISCORD_ID,
    )
    unban = mocker.patch(
        "moddingway_api.routes.banforms_routes.unban_user", new=mocker.AsyncMock()
    )

    response = client.patch(
        "/banforms", json={"form_id": "4", "approval": True, "approver_id": "9"}
    )

    assert response.json() == {"detail": "Form 4 has been updated"}
    update.assert_called_once_with("4", True, "9")
    unban.assert_awaited_once_with(DISCORD_ID)


def test_denying_a_form_does_not_unban(client, mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.banforms_database.get_ban_form", return_value=FORM
    )
    mocker.patch(
        "moddingway.database.banforms_database.update_form", return_value=(False, 4)
    )
    unban = mocker.patch(
        "moddingway_api.routes.banforms_routes.unban_user", new=mocker.AsyncMock()
    )

    client.patch(
        "/banforms", json={"form_id": "4", "approval": False, "approver_id": "9"}
    )

    unban.assert_not_awaited()


def test_updating_a_missing_form_is_not_found(client, mocker: MockerFixture):
    mocker.patch(
        "moddingway.database.banforms_database.get_ban_form", return_value=None
    )

    assert (
        client.patch(
            "/banforms", json={"form_id": "4", "approval": True, "approver_id": "9"}
        ).status_code
        == HTTPStatus.NOT_FOUND
    )


def test_lifespan_connects_and_disconnects_the_database(mocker: MockerFixture):
    db = mocker.patch("moddingway_api.main.DatabaseConnection").return_value

    with TestClient(api_main.app) as c:
        assert c.get("/health").status_code == HTTPStatus.OK
        db.connect.assert_called_once()
        db.create_tables.assert_called_once()
        db.disconnect.assert_not_called()

    db.disconnect.assert_called_once()
