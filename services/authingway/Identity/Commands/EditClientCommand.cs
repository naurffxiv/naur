// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Responses;
using Naur.Authingway.Discord.Preconditions;
using Naur.Authingway.Identity.Services;
using Naur.Authingway.Identity.Responses;
using Naur.Authingway.Identity.Modals;
using NetCord;
using NetCord.Rest;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Identity.Commands;

/// <summary>
/// Opens the form for editing an OpenIddict client.
/// </summary>
/// <param name="clients">The client registration service.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed class EditClientCommand(ClientRegistrationService clients) : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// Shows the edit form prefilled with the client's current settings.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <returns>A task representing the operation.</returns>
    [SlashCommand("client-edit", "Edit an OpenID Connect client.", Contexts = [InteractionContextType.Guild])]
    public async Task ExecuteAsync(
        [SlashCommandParameter(Name = "client_id", Description = "The client ID.")] string clientId)
    {
        var client = await clients.FindAsync(clientId);

        if (client is null)
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.NotFound(clientId)));
            return;
        }

        if (!ClientModal.TryCreateCustomId(ClientModal.EditCustomId, client.ClientId, out var customId))
        {
            await RespondAsync(EphemeralResponse.Message("This client ID cannot be edited from Discord."));
            return;
        }

        await RespondAsync(InteractionCallback.Modal(ClientModal.ForEdit(customId, client.Settings)));
    }
}
