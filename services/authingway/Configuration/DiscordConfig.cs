// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Configuration;

/// <summary>
/// Maps the named Discord roles and channel groups to their identifiers.
/// </summary>
/// <remarks>
/// Bound from the <c>Discord</c> configuration section, for example
/// <c>Discord:Roles:Admin = 123</c> and <c>Discord:Channels:Developer:0 = 456</c>.
/// </remarks>
public class DiscordConfig
{
    /// <summary>
    /// Gets or sets the Discord role identifier for each named role.
    /// </summary>
    /// <value>Role identifiers keyed by role; a role with no entry matches no member.</value>
    public Dictionary<DiscordRole, ulong?> Roles { get; set; } = [];

    /// <summary>
    /// Gets or sets the Discord channel identifiers for each named channel group.
    /// </summary>
    /// <value>Channel identifiers keyed by channel group.</value>
    public Dictionary<DiscordChannelGroup, List<ulong>> Channels { get; set; } = [];
}
