# TODO: look into using psycopg to handle model mapping for us
# https://www.psycopg.org/docs/advanced.html#adapting-new-python-types-to-sql-syntax
class User(object):
    def __init__(self, user_id, discord_user_id, discord_guild_id, is_mod):
        self.user_id = user_id
        self.discord_user_id = discord_user_id
        self.discord_guild_id = discord_guild_id
        self.is_mod = is_mod
