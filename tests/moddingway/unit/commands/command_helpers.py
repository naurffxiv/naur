"""Helpers for driving slash commands in tests."""

import discord

from moddingway import constants


async def run(bot, name, interaction, **kwargs):
    command = bot.tree.get_command(name)
    assert command is not None, f"/{name} is not registered"
    await command.callback(interaction, **kwargs)


def reply(interaction) -> str:
    """The final text of the command's ephemeral reply."""
    return interaction.original.edit.await_args.kwargs["content"]


def sent(interaction) -> str:
    """The text of the command's direct response."""
    return interaction.response.send_message.await_args.args[0]


def logged(interaction) -> discord.Embed:
    return interaction.log_channel.send.await_args.kwargs["embed"]


def target(mocker, id=12345, mod=False, member=True, role=None):
    spec = discord.Member if member else discord.User
    user = mocker.Mock(
        spec=spec,
        id=id,
        mention=f"<@{id}>",
        display_name="target",
        display_avatar=mocker.Mock(url="u"),
    )
    if role is None:
        role = constants.Role.MOD if mod else constants.Role.VERIFIED
    discord_role = mocker.Mock()
    discord_role.name = role.value
    user.roles = [discord_role]
    return user
