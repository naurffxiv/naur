from moddingway.database import users_database


def test_database__users_are_saved_and_read_back(query):
    created = users_database.add_user(12345)

    found = users_database.get_user(12345)

    assert found is not None
    assert found.user_id == created.user_id
    assert query("SELECT discordUserID FROM users") == [("12345",)]


def test_database__is_emptied_between_tests(query):
    assert query("SELECT * FROM users") == []
