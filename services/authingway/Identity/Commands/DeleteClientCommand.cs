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
/// Opens the confirmation form for deleting an OpenIddict client.
/// </summary>
/// <param name="clients">The client registration service.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed class DeleteClientCommand(ClientRegistrationService clients) : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// Shows the delete confirmation form.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <returns>A task representing the operation.</returns>
    [SlashCommand("client-delete", "Delete an OpenID Connect client.", Contexts = [InteractionContextType.Guild])]
    public async Task ExecuteAsync(
        [SlashCommandParameter(Name = "client_id", Description = "The client ID.")] string clientId)
    {
        var client = await clients.FindAsync(clientId);

        if (client is null)
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.NotFound(clientId)));
            return;
        }

        if (!ClientModal.TryCreateCustomId(DeleteClientModal.CustomId, client.ClientId, out var customId))
        {
            await RespondAsync(EphemeralResponse.Message("This client ID cannot be deleted from Discord."));
            return;
        }

        await RespondAsync(InteractionCallback.Modal(
            DeleteClientModal.Build(customId, client.ClientId, client.Settings.DisplayName)));
    }
}
