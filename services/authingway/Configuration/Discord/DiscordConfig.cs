// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.ComponentModel.DataAnnotations;
using Naur.Authingway.Configuration.Validators;

namespace Naur.Authingway.Configuration.Discord;

/// <summary>
/// The discord configuration for authingway.
/// </summary>
public class DiscordConfig
{
    /// <summary>
    /// Gets or sets the bot token from the Discord section shared with the gateway.
    /// </summary>
    public string? Token { get; set; }

    /// <summary>
    /// The guild this authingway instance is configured for.
    /// </summary>
    [Range(1, ulong.MaxValue, ErrorMessage = "A valid guild id must be set.")]
    public ulong GuildId { get; set; }

    /// <summary>
    /// Gets or sets the Discord role identifier for each named role.
    /// </summary>
    [ValidRoles]
    public Dictionary<DiscordRole, ulong> Roles { get; set; } = [];

    /// <summary>
    /// Gets or sets the Discord channel identifiers for each named channel group.
    /// </summary>
    [ValidChannels]
    public Dictionary<DiscordChannelGroup, HashSet<ulong>> Channels { get; set; } = [];
}
