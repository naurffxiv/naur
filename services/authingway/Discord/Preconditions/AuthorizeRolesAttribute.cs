// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using NetCord;
using NetCord.Services;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Discord.Preconditions;

/// <summary>
/// Allows an application command for a member holding any of the specified configured roles.
/// An empty list or a role without a configured, nonzero identifier grants no access.
/// </summary>
/// <param name="roles">The allowed roles. Guild and channel restrictions are applied separately.</param>
public sealed class AuthorizeRolesAttribute(DiscordRole[] roles) : AuthorizeRolesAttribute<ApplicationCommandContext>(roles);

/// <summary>
/// Allows an interaction for a member holding any of the specified configured roles.
/// An empty list or a role without a configured, nonzero identifier grants no access.
/// </summary>
/// <typeparam name="TContext">The interaction context the precondition applies to.</typeparam>
/// <param name="roles">The allowed roles. Guild and channel restrictions are applied separately.</param>
public abstract class AuthorizeRolesAttribute<TContext>(DiscordRole[] roles) : PreconditionAttribute<TContext>
    where TContext : IInteractionContext
{
    /// <inheritdoc/>
    public override ValueTask<PreconditionResult> EnsureCanExecuteAsync(TContext context, IServiceProvider? serviceProvider)
    {
        ArgumentNullException.ThrowIfNull(serviceProvider);

        var config = serviceProvider.GetRequiredService<IOptions<DiscordConfig>>().Value;

        var permitted = context.Interaction.User is GuildUser user
            && roles.Any(role => config.Roles.TryGetValue(role, out var id) && id != 0 && user.RoleIds.Contains(id));

        return new(permitted
            ? PreconditionResult.Success
            : PreconditionResult.Fail(context.Interaction.GuildId is null
                ? "This command is not available in DMs."
                : "You do not have a required role to use this command."));
    }
}
