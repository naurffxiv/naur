from datetime import datetime

from moddingway.database import banforms_database, users_database
from moddingway.database.models import BanForm


def _form(user_id, reason="please", when=datetime(2026, 1, 1)):
    return banforms_database.add_form(
        BanForm(user_id=user_id, reason=reason, submission_timestamp=when)
    )


def test_add_and_get_form():
    user = users_database.add_user(12345)
    form_id = _form(user.user_id)

    form = banforms_database.get_ban_form(form_id)

    assert form == BanForm(
        form_id=form_id,
        user_id=user.user_id,
        reason="please",
        submission_timestamp=datetime(2026, 1, 1),
    )
    assert banforms_database.get_ban_form(999) is None
    assert banforms_database.get_user_from_form(form_id) == "12345"
    assert banforms_database.get_user_from_form(999) is None
    assert banforms_database.get_form_count() == 1


def test_get_ban_forms_pages_through_forms():
    user = users_database.add_user(1)
    approver = users_database.add_user(2)
    first = _form(user.user_id, "first", datetime(2026, 1, 1))
    second = _form(user.user_id, "second", datetime(2026, 1, 2))
    banforms_database.update_form(second, True, approver.user_id)

    forms = banforms_database.get_ban_forms(10, 0)

    assert forms == [
        BanForm(
            form_id=first,
            user_id=user.user_id,
            reason="first",
            submission_timestamp=datetime(2026, 1, 1),
        ),
        BanForm(
            form_id=second,
            user_id=user.user_id,
            reason="second",
            approval=True,
            approved_by=approver.user_id,
            submission_timestamp=datetime(2026, 1, 2),
        ),
    ]
    assert banforms_database.get_ban_forms(10, 2) == []


def test_update_form_records_the_decision():
    user = users_database.add_user(1)
    approver = users_database.add_user(2)
    form_id = _form(user.user_id)

    assert banforms_database.update_form(form_id, False, approver.user_id) == (
        False,
        form_id,
    )
    form = banforms_database.get_ban_form(form_id)
    assert (form.approval, form.approved_by) == (False, approver.user_id)
    assert banforms_database.update_form(999, True, approver.user_id) is None
