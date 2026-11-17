// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration;
using NetCord;
using NetCord.Services;

namespace Naur.Authingway.Discord;

/// <summary>
/// Allows a command only for members holding one of the given roles, in one of the given channel groups.
/// </summary>
/// <typeparam name="TContext">The command context type. Must expose the channel and user.</typeparam>
/// <remarks>
/// Roles and channel groups resolve to identifiers through <see cref="DiscordConfig"/>. A property left empty does
/// not restrict that dimension, but a role or group with no configured identifier matches nothing, so missing
/// configuration keeps the command closed.
/// </remarks>
/// <example>
/// <code>
/// [RequireDiscordRoles&lt;ApplicationCommandContext&gt;(
///     Roles = [DiscordRole.Admin, DiscordRole.Developer],
///     Channels = [DiscordChannelGroup.Admin, DiscordChannelGroup.Developer])]
/// </code>
/// </example>
public class RequireDiscordRolesAttribute<TContext> : PreconditionAttribute<TContext>
    where TContext : IChannelContext, IUserContext
{
    /// <summary>
    /// Gets or sets the roles the invoking member must hold at least one of.
    /// </summary>
    /// <value>The roles, or an empty array to allow any role.</value>
    public DiscordRole[] Roles { get; set; } = [];

    /// <summary>
    /// Gets or sets the channel groups the command must be invoked in one of.
    /// </summary>
    /// <value>The channel groups, or an empty array to allow any channel.</value>
    public DiscordChannelGroup[] Channels { get; set; } = [];

    /// <inheritdoc/>
    public override ValueTask<PreconditionResult> EnsureCanExecuteAsync(TContext context, IServiceProvider? serviceProvider)
    {
        ArgumentNullException.ThrowIfNull(serviceProvider);

        var config = serviceProvider.GetRequiredService<IOptions<DiscordConfig>>().Value;

        return new(IsPermitted(config, context.Channel?.Id, (context.User as GuildUser)?.RoleIds)
            ? PreconditionResult.Success
            : PreconditionResult.Fail("This command is not available here."));
    }

    /// <summary>
    /// Determines whether an invocation from the given channel by a member holding the given roles is permitted.
    /// </summary>
    /// <param name="config">The configured role and channel identifiers. Cannot be null.</param>
    /// <param name="channelId">The identifier of the channel the command was invoked from, or <see langword="null"/> when unknown.</param>
    /// <param name="roleIds">The roles held by the invoking guild member, or <see langword="null"/> outside a guild.</param>
    /// <returns><see langword="true"/> when the invocation is permitted; otherwise, <see langword="false"/>.</returns>
    public bool IsPermitted(DiscordConfig config, ulong? channelId, IEnumerable<ulong>? roleIds)
    {
        ArgumentNullException.ThrowIfNull(config);

        if (Roles.Length > 0)
        {
            var allowedRoles = Roles
                .Select(role => config.Roles.GetValueOrDefault(role))
                .OfType<ulong>()
                .Where(id => id != 0)
                .ToHashSet();

            if (roleIds is null || !roleIds.Any(allowedRoles.Contains))
            {
                return false;
            }
        }

        if (Channels.Length > 0)
        {
            var allowedChannels = Channels
                .SelectMany(group => config.Channels.GetValueOrDefault(group) ?? [])
                .Where(id => id != 0)
                .ToHashSet();

            if (channelId is not { } id || !allowedChannels.Contains(id))
            {
                return false;
            }
        }

        return true;
    }
}
