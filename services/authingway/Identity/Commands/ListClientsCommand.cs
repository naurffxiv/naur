// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Text;
using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Responses;
using Naur.Authingway.Discord.Preconditions;
using Naur.Authingway.Identity.Services;
using NetCord;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Identity.Commands;

/// <summary>
/// Lists the registered OpenIddict clients.
/// </summary>
/// <param name="clients">The client registration service.</param>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed class ListClientsCommand(ClientRegistrationService clients) : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// Shows a page of clients.
    /// </summary>
    /// <param name="page">The one-based page number.</param>
    /// <returns>A task representing the operation.</returns>
    [SlashCommand("client-list", "List the registered OpenID Connect clients.", Contexts = [InteractionContextType.Guild])]
    public async Task ExecuteAsync(
        [SlashCommandParameter(Name = "page", Description = "The page to show.", MinValue = 1)] int page = 1)
    {
        var result = await clients.ListAsync(page);

        if (result.Total == 0)
        {
            await RespondAsync(EphemeralResponse.Message("No clients are registered. Use `/client-create` to add one."));
            return;
        }

        if (result.Clients.Count == 0)
        {
            await RespondAsync(EphemeralResponse.Message($"Page {page} is empty. There are {result.PageCount} page(s)."));
            return;
        }

        var builder = new StringBuilder($"**Clients** (page {result.Page} of {result.PageCount}, {result.Total} total)\n");

        foreach (var client in result.Clients)
        {
            builder.AppendLine($"- {EphemeralResponse.Escape(client.DisplayName ?? "(unnamed)")}: `{client.ClientId}`");
        }

        await RespondAsync(EphemeralResponse.Message(builder.ToString()));
    }
}
