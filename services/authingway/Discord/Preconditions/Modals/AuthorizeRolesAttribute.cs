// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using NetCord.Services.ComponentInteractions;

namespace Naur.Authingway.Discord.Preconditions.Modals;

/// <summary>
/// Allows a modal submission for a member holding any of the specified configured roles.
/// An empty list or a role without a configured, nonzero identifier grants no access.
/// </summary>
/// <param name="roles">The allowed roles. Guild and channel restrictions are applied separately.</param>
public sealed class AuthorizeRolesAttribute(DiscordRole[] roles) : AuthorizeRolesAttribute<ModalInteractionContext>(roles);
