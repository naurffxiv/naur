// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;

namespace Naur.Authingway.Discord;

/// <summary>
/// Decides whether a migration command invocation is permitted.
/// </summary>
/// <param name="options">The configured Discord identifiers. Cannot be null.</param>
public sealed class MigrationCommandAuthorizer(IOptions<DiscordOptions> options)
{
    /// <summary>
    /// Determines whether an invocation from the given guild and channel by a user holding the given roles is permitted.
    /// </summary>
    /// <param name="guildId">The identifier of the guild the command was invoked from, or <see langword="null"/> outside a guild.</param>
    /// <param name="channelId">The identifier of the channel the command was invoked from.</param>
    /// <param name="roleIds">The identifiers of the roles held by the invoking user. Cannot be null.</param>
    /// <returns><see langword="true"/> when the invocation is permitted; otherwise, <see langword="false"/>.</returns>
    /// <remarks>
    /// Authorization fails when any identifier is unconfigured, so an incomplete configuration permits nothing.
    /// </remarks>
    public bool IsAuthorized(ulong? guildId, ulong channelId, IEnumerable<ulong> roleIds)
    {
        ArgumentNullException.ThrowIfNull(roleIds);

        var configured = options.Value;

        if (configured.GuildId is not ({ } configuredGuildId and not 0)
            || configured.MigrationChannelId is not ({ } migrationChannelId and not 0)
            || configured.DeveloperRoleId is not ({ } developerRoleId and not 0))
        {
            return false;
        }

        if (guildId != configuredGuildId || channelId != migrationChannelId)
        {
            return false;
        }

        return roleIds.Contains(developerRoleId);
    }
}
