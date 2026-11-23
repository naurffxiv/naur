// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Responses;
using Naur.Authingway.Discord.Preconditions.Modals;
using Naur.Authingway.Identity.Services;
using Naur.Authingway.Identity.Responses;
using NetCord.Services.ComponentInteractions;

namespace Naur.Authingway.Identity.Modals;

/// <summary>
/// Handles submissions of the edit client form.
/// </summary>
/// <param name="clients">The client registration service.</param>
/// <param name="logger">The logger.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed partial class EditClientModal(ClientRegistrationService clients, ILogger<EditClientModal> logger)
    : ComponentInteractionModule<ModalInteractionContext>
{
    /// <summary>
    /// Updates the client's settings, keeping its secret.
    /// </summary>
    /// <param name="clientId">The client identifier carried by the form.</param>
    /// <returns>A task representing the operation.</returns>
    [ComponentInteraction(ClientModal.EditCustomId)]
    public async Task ExecuteAsync(string clientId)
    {
        var result = ClientModal.Read(Context.Components);

        if (result.Settings is null)
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.Invalid(result.Errors)));
            return;
        }

        var credentials = await clients.UpdateAsync(clientId, result.Settings);

        if (credentials is null)
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.NotFound(clientId)));
            return;
        }

        LogUpdated(clientId, Context.User.Id);

        var updated = $"Updated **{EphemeralResponse.Escape(result.Settings.DisplayName)}** (`{clientId}`).";

        await RespondAsync(EphemeralResponse.Message(
            credentials.ClientSecret is not null ? $"{updated}\nThe client is now confidential.\n{ClientMessages.Secret(credentials.ClientSecret)}"
            : result.Settings.IsPublic ? $"{updated} This is a public client, so it has no secret."
            : $"{updated} The client secret is unchanged."));
    }

    [LoggerMessage(LogLevel.Information, "Client {ClientId} was updated by Discord user {UserId}.")]
    private partial void LogUpdated(string clientId, ulong userId);
}
