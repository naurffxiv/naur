from .autounexile import autounexile_users
from .forum_automod import autodelete_posts, autodelete_threads
from .strike_decrement import decrement_strikes

WORKERS = (
    autounexile_users,
    autodelete_threads,
    decrement_strikes,
    autodelete_posts,
)


def start_tasks(bot):
    for worker in WORKERS:
        if worker.is_running():
            continue
        worker.clear_exception_types()
        worker.add_exception_type(Exception)
        worker.start(bot)
