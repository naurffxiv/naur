// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;
using NetCord;
using NetCord.Services;

namespace Naur.Authingway.Discord;

/// <summary>
/// Allows a command only in the channel and for the role configured by a named <see cref="ChannelRoleRestrictionOptions"/>.
/// </summary>
/// <typeparam name="TContext">The command context type. Must expose the channel and user.</typeparam>
/// <param name="restriction">The name of the restriction to enforce. Cannot be null.</param>
/// <example>
/// <code>
/// [RequireChannelRole&lt;ApplicationCommandContext&gt;("Migration")]
/// </code>
/// </example>
public class RequireChannelRoleAttribute<TContext>(string restriction) : PreconditionAttribute<TContext>
    where TContext : IChannelContext, IUserContext
{
    /// <summary>
    /// Gets the name of the restriction this attribute enforces.
    /// </summary>
    /// <value>The restriction name.</value>
    public string Restriction { get; } = restriction;

    /// <inheritdoc/>
    /// <remarks>
    /// The guild and roles come from the invoking user, which is a <see cref="GuildUser"/> only inside a guild, so
    /// invocations outside a guild are refused.
    /// </remarks>
    public override ValueTask<PreconditionResult> EnsureCanExecuteAsync(TContext context, IServiceProvider? serviceProvider)
    {
        ArgumentNullException.ThrowIfNull(serviceProvider);

        var options = serviceProvider.GetRequiredService<IOptionsMonitor<ChannelRoleRestrictionOptions>>().Get(Restriction);
        var guildUser = context.User as GuildUser;

        return new(options.Permits(guildUser?.GuildId, context.Channel?.Id, guildUser?.RoleIds ?? [])
            ? PreconditionResult.Success
            : PreconditionResult.Fail("This command is not available here."));
    }
}
