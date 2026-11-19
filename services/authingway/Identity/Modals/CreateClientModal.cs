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
/// Handles submissions of the create client form.
/// </summary>
/// <param name="clients">The client registration service.</param>
/// <param name="logger">The logger.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed partial class CreateClientModal(ClientRegistrationService clients, ILogger<CreateClientModal> logger)
    : ComponentInteractionModule<ModalInteractionContext>
{
    /// <summary>
    /// Creates the client and reveals its secret once.
    /// </summary>
    /// <returns>A task representing the operation.</returns>
    [ComponentInteraction(ClientModal.CreateCustomId)]
    public async Task ExecuteAsync()
    {
        var result = ClientModal.Read(Context.Components);

        if (result.Settings is null)
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.Invalid(result.Errors)));
            return;
        }

        var credentials = await clients.CreateAsync(result.Settings);

        LogCreated(credentials.ClientId, Context.User.Id);

        await RespondAsync(EphemeralResponse.Message(
            $"Created **{EphemeralResponse.Escape(result.Settings.DisplayName)}**.\n" +
            $"Client ID: `{credentials.ClientId}`\n" +
            $"Client secret: `{credentials.ClientSecret}`\n" +
            "Copy the secret now. It is stored hashed and cannot be shown again."));
    }

    [LoggerMessage(LogLevel.Information, "Client {ClientId} was created by Discord user {UserId}.")]
    private partial void LogCreated(string clientId, ulong userId);
}
