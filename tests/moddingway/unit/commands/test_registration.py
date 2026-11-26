from moddingway import util
from moddingway.commands.announcement_commands import create_announcement_commands
from moddingway.commands.ban_commands import create_ban_commands
from moddingway.commands.exile_commands import create_exile_commands
from moddingway.commands.forum_commands import create_forum_commands
from moddingway.commands.note_commands import create_note_commands
from moddingway.commands.send_message_commands import create_send_message_commands
from moddingway.commands.slowmode_commands import create_slowmode_commands
from moddingway.commands.strikes_command import create_strikes_commands
from moddingway.commands.warning_commands import create_warning_commands

ADMIN_ONLY = {
    "draft_announcement",
    "publish_announcement",
    "list_announcements",
    "show_announcement",
}
EVERYONE = {"roulette"}


def test_every_command_is_registered_with_the_right_permission_check(bot):
    for create in (
        create_announcement_commands,
        create_ban_commands,
        create_exile_commands,
        create_forum_commands,
        create_note_commands,
        create_send_message_commands,
        create_slowmode_commands,
        create_strikes_commands,
        create_warning_commands,
    ):
        create(bot)

    commands = {c.name: c for c in bot.tree.get_commands()}
    assert set(commands) == {
        "draft_announcement",
        "publish_announcement",
        "list_announcements",
        "show_announcement",
        "ban",
        "unexile",
        "exile",
        "roulette",
        "view_logged_exiles",
        "view_active_exiles",
        "delete_exile",
        "run_auto_unexile",
        "run_automod",
        "add_note",
        "view_notes",
        "delete_note",
        "edit_note",
        "send_message",
        "set_slowmode",
        "add_strike",
        "view_strikes",
        "delete_strike",
        "run_strike_decrement",
        "warn",
        "view_warnings",
    }
    for name, command in commands.items():
        predicates = [getattr(check, "__wrapped__", check) for check in command.checks]
        if name in EVERYONE:
            assert (
                util.is_user_moderator not in predicates
                and util.is_user_admin not in predicates
            )
        elif name in ADMIN_ONLY:
            assert util.is_user_admin in predicates, name
        else:
            assert util.is_user_moderator in predicates, name


def test_roulette_has_a_daily_cooldown(bot):
    create_exile_commands(bot)

    cooldown = bot.tree.get_command("roulette").checks[0]
    assert "cooldown" in getattr(cooldown, "__qualname__", "")
