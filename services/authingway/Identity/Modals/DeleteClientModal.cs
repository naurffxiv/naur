// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Responses;
using Naur.Authingway.Discord.Extensions;
using Naur.Authingway.Discord.Preconditions.Modals;
using Naur.Authingway.Identity.Services;
using Naur.Authingway.Identity.Responses;
using NetCord;
using NetCord.Rest;
using NetCord.Services.ComponentInteractions;

namespace Naur.Authingway.Identity.Modals;

/// <summary>
/// The confirmation form shown before deleting a client, and the handler for its submissions.
/// </summary>
/// <param name="clients">The client registration service.</param>
/// <param name="logger">The logger.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed partial class DeleteClientModal(ClientRegistrationService clients, ILogger<DeleteClientModal> logger)
    : ComponentInteractionModule<ModalInteractionContext>
{
    /// <summary>
    /// The custom identifier prefix of the form; the client identifier follows it.
    /// </summary>
    public const string CustomId = "client-delete";

    /// <summary>
    /// The confirmation input.
    /// </summary>
    public const string ConfirmationInput = "confirm_client_id";

    private const int MaxDescriptionLength = 100;

    /// <summary>
    /// Builds the confirmation form.
    /// </summary>
    /// <param name="customId">The custom identifier carrying the client identifier.</param>
    /// <param name="clientId">The client identifier to confirm.</param>
    /// <param name="displayName">The client display name.</param>
    /// <returns>The modal.</returns>
    public static ModalProperties Build(string customId, string clientId, string displayName)
    {
        ArgumentNullException.ThrowIfNull(clientId);

        var description = $"Permanently deletes {displayName}.";

        if (description.Length > MaxDescriptionLength)
        {
            description = description[..(MaxDescriptionLength - 1)] + "…";
        }

        var confirmation = new TextInputProperties(ConfirmationInput, TextInputStyle.Short)
            .WithMinLength(clientId.Length)
            .WithMaxLength(clientId.Length)
            .WithRequired(true)
            .WithPlaceholder(clientId);

        return new ModalProperties(customId, "Delete client",
        [
            new LabelProperties("Type the client ID to confirm", confirmation)
                .WithDescription(description)
        ]);
    }

    /// <summary>
    /// Determines whether the submitted confirmation matches the client identifier exactly.
    /// </summary>
    /// <param name="components">The submitted components.</param>
    /// <param name="clientId">The client identifier being deleted.</param>
    /// <returns>True when the deletion is confirmed.</returns>
    public static bool IsConfirmed(IEnumerable<IModalComponent> components, string clientId)
    {
        return string.Equals(components.GetText(ConfirmationInput).Trim(), clientId, StringComparison.Ordinal);
    }

    /// <summary>
    /// Deletes the client when the typed identifier matches.
    /// </summary>
    /// <param name="clientId">The client identifier carried by the form.</param>
    /// <returns>A task representing the operation.</returns>
    [ComponentInteraction(CustomId)]
    public async Task ExecuteAsync(string clientId)
    {
        if (!IsConfirmed(Context.Components, clientId))
        {
            await RespondAsync(EphemeralResponse.Message("The client ID did not match. Nothing was deleted."));
            return;
        }

        if (!await clients.DeleteAsync(clientId))
        {
            await RespondAsync(EphemeralResponse.Message(ClientMessages.NotFound(clientId)));
            return;
        }

        LogDeleted(clientId, Context.User.Id);

        await RespondAsync(EphemeralResponse.Message($"Deleted client `{clientId}`."));
    }

    [LoggerMessage(LogLevel.Information, "Client {ClientId} was deleted by Discord user {UserId}.")]
    private partial void LogDeleted(string clientId, ulong userId);
}
