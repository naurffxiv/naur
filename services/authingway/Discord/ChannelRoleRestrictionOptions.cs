// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Discord;

/// <summary>
/// Restricts a Discord command to one channel of one guild, for members holding one role.
/// </summary>
/// <remarks>
/// Each restriction is a named options instance bound from <c>Discord:Restrictions:{name}</c>, so commands can be
/// restricted independently.
/// </remarks>
public sealed class ChannelRoleRestrictionOptions
{
    /// <summary>
    /// The configuration section that holds one subsection per named restriction.
    /// </summary>
    public const string SectionName = "Discord:Restrictions";

    /// <summary>
    /// Gets or sets the identifier of the guild the command is restricted to.
    /// </summary>
    /// <value>The guild identifier, or <see langword="null"/> or zero when no guild has been configured.</value>
    public ulong? GuildId { get; set; }

    /// <summary>
    /// Gets or sets the identifier of the only channel in which the command is accepted.
    /// </summary>
    /// <value>The channel identifier, or <see langword="null"/> or zero when no channel has been configured.</value>
    public ulong? ChannelId { get; set; }

    /// <summary>
    /// Gets or sets the identifier of the role whose members may run the command.
    /// </summary>
    /// <value>The role identifier, or <see langword="null"/> or zero when no role has been configured.</value>
    public ulong? RoleId { get; set; }

    /// <summary>
    /// Determines whether an invocation from the given guild and channel by a user holding the given roles is permitted.
    /// </summary>
    /// <param name="guildId">The identifier of the guild the command was invoked from, or <see langword="null"/> outside a guild.</param>
    /// <param name="channelId">The identifier of the channel the command was invoked from, or <see langword="null"/> when unknown.</param>
    /// <param name="roleIds">The identifiers of the roles held by the invoking user. Cannot be null.</param>
    /// <returns><see langword="true"/> when the invocation is permitted; otherwise, <see langword="false"/>.</returns>
    /// <remarks>
    /// Nothing is permitted while any identifier is unconfigured, so an incomplete configuration disables the command.
    /// </remarks>
    public bool Permits(ulong? guildId, ulong? channelId, IEnumerable<ulong> roleIds)
    {
        ArgumentNullException.ThrowIfNull(roleIds);

        if (GuildId is not ({ } requiredGuildId and not 0)
            || ChannelId is not ({ } requiredChannelId and not 0)
            || RoleId is not ({ } requiredRoleId and not 0))
        {
            return false;
        }

        return guildId == requiredGuildId
            && channelId == requiredChannelId
            && roleIds.Contains(requiredRoleId);
    }
}
