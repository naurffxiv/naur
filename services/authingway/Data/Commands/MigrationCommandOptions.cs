// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Data.Commands;

/// <summary>
/// Holds the Discord identifiers that restrict who may run privileged commands and where.
/// </summary>
public sealed class MigrationCommandOptions
{
    /// <summary>
    /// The configuration section these options bind from.
    /// </summary>
    public const string SectionName = "Discord";

    /// <summary>
    /// Gets or sets the identifier of the guild whose channel and role restrict privileged commands.
    /// </summary>
    /// <value>The guild identifier, or <see langword="null"/> or zero when no guild has been configured.</value>
    public ulong? GuildId { get; set; }

    /// <summary>
    /// Gets or sets the identifier of the only channel in which the migration command is accepted.
    /// </summary>
    /// <value>The channel identifier, or <see langword="null"/> or zero when no channel has been configured.</value>
    public ulong? MigrationChannelId { get; set; }

    /// <summary>
    /// Gets or sets the identifier of the role whose members may run the migration command.
    /// </summary>
    /// <value>The role identifier, or <see langword="null"/> or zero when no role has been configured.</value>
    public ulong? DeveloperRoleId { get; set; }
}
