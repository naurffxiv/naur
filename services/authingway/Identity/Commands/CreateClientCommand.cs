// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Preconditions;
using Naur.Authingway.Identity.Modals;
using NetCord;
using NetCord.Rest;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Identity.Commands;

/// <summary>
/// Opens the form for registering a new OpenIddict client.
/// </summary>
[AuthorizeGuild]
[AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
public sealed class CreateClientCommand : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// Shows the create client form.
    /// </summary>
    /// <returns>A task representing the operation.</returns>
    [SlashCommand("client-create", "Register a new OpenID Connect client.", Contexts = [InteractionContextType.Guild])]
    public async Task ExecuteAsync()
    {
        await RespondAsync(InteractionCallback.Modal(ClientModal.ForCreate()));
    }
}
