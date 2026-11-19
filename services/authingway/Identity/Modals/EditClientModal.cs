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

        if (!await clients.UpdateAsync(clientId, result.Settings))
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.NotFound(clientId)));
            return;
        }

        LogUpdated(clientId, Context.User.Id);

        await RespondAsync(EphemeralResponse.Message(
            $"Updated **{EphemeralResponse.Escape(result.Settings.DisplayName)}** (`{clientId}`). The client secret is unchanged."));
    }

    [LoggerMessage(LogLevel.Information, "Client {ClientId} was updated by Discord user {UserId}.")]
    private partial void LogUpdated(string clientId, ulong userId);
}
