// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Configuration;

/// <summary>
/// Names the groups of Discord channels that commands can be restricted to; each maps to channel identifiers in <see cref="DiscordConfig.Channels"/>.
/// </summary>
public enum DiscordChannelGroup
{
    /// <summary>
    /// The administrator channels.
    /// </summary>
    Admin,

    /// <summary>
    /// The developer channels.
    /// </summary>
    Developer
}
