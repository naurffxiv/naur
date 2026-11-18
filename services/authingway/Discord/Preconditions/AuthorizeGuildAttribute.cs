// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using NetCord.Services;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Discord.Preconditions;

/// <summary>
/// Allows an application command only in the guild configured by <see cref="DiscordConfig.GuildId"/>.
/// </summary>
public sealed class AuthorizeGuildAttribute : PreconditionAttribute<ApplicationCommandContext>
{
    /// <inheritdoc/>
    public override ValueTask<PreconditionResult> EnsureCanExecuteAsync(ApplicationCommandContext context, IServiceProvider? serviceProvider)
    {
        ArgumentNullException.ThrowIfNull(serviceProvider);

        var config = serviceProvider.GetRequiredService<IOptions<DiscordConfig>>().Value;

        return new(config.GuildId != 0 && context.Interaction.GuildId == config.GuildId
            ? PreconditionResult.Success
            : PreconditionResult.Fail(context.Interaction.GuildId is null
                ? "This command is not available in DMs."
                : "This command is not available in this guild."));
    }
}
