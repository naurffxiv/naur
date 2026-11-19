// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using NetCord.Services.ComponentInteractions;

namespace Naur.Authingway.Discord.Preconditions.Modals;

/// <summary>
/// Allows a modal submission only in the guild configured by <see cref="DiscordConfig.GuildId"/>.
/// </summary>
public sealed class AuthorizeGuildAttribute : AuthorizeGuildAttribute<ModalInteractionContext>;
