// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using NetCord.Services;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Discord.Preconditions;

/// <summary>
/// Allows an application command in any channel belonging to the specified configured groups.
/// An empty list or a group without configured, nonzero identifiers grants no access.
/// </summary>
/// <param name="channels">The allowed channel groups. Guild and role restrictions are applied separately.</param>
public sealed class AuthorizeChannelsAttribute(DiscordChannelGroup[] channels) : PreconditionAttribute<ApplicationCommandContext>
{
    /// <inheritdoc/>
    public override ValueTask<PreconditionResult> EnsureCanExecuteAsync(ApplicationCommandContext context, IServiceProvider? serviceProvider)
    {
        ArgumentNullException.ThrowIfNull(serviceProvider);

        var config = serviceProvider.GetRequiredService<IOptions<DiscordConfig>>().Value;

        var permitted = context.Channel?.Id is > 0 and var id
            && channels.Any(group => config.Channels.TryGetValue(group, out var ids) && ids.Contains(id));

        return new(permitted
            ? PreconditionResult.Success
            : PreconditionResult.Fail(context.Interaction.GuildId is null
                ? "This command is not available in DMs."
                : "This command is not available in this channel."));
    }
}
