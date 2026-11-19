// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Responses;
using Naur.Authingway.Discord.Preconditions;
using Naur.Authingway.Identity.Services;
using Naur.Authingway.Identity.Responses;
using NetCord;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Identity.Commands;

/// <summary>
/// Shows the settings of an OpenIddict client. Secrets are never shown.
/// </summary>
/// <param name="clients">The client registration service.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed class ViewClientCommand(ClientRegistrationService clients) : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// Shows the client's settings.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <returns>A task representing the operation.</returns>
    [SlashCommand("client-view", "Show an OpenID Connect client's settings.", Contexts = [InteractionContextType.Guild])]
    public async Task ExecuteAsync(
        [SlashCommandParameter(Name = "client_id", Description = "The client ID.")] string clientId)
    {
        var client = await clients.FindAsync(clientId);

        await RespondAsync(EphemeralResponse.Message(client is null
            ? ClientMessages.NotFound(clientId)
            : ClientMessages.Describe(client)));
    }
}
