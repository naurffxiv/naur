// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Configuration;

/// <summary>
/// Names the Discord roles that commands can be restricted to; each maps to a role identifier in <see cref="DiscordConfig.Roles"/>.
/// </summary>
public enum DiscordRole
{
    /// <summary>
    /// The server administrator role.
    /// </summary>
    Admin,

    /// <summary>
    /// The developer role.
    /// </summary>
    Developer
}
