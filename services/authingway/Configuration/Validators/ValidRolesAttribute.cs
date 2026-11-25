// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.ComponentModel.DataAnnotations;
using Naur.Authingway.Configuration.Discord;

namespace Naur.Authingway.Configuration.Validators;

/// <summary>
/// Validates that Discord roles have known names and positive identifiers.
/// </summary>
[AttributeUsage(AttributeTargets.Property | AttributeTargets.Field | AttributeTargets.Parameter)]
public sealed class ValidRolesAttribute : ValidationAttribute
{
    /// <summary>
    /// Initializes a new instance of the <see cref="ValidRolesAttribute"/> class.
    /// </summary>
    public ValidRolesAttribute()
        : base("Roles must contain only known Discord roles with positive identifiers.")
    {
    }

    /// <inheritdoc/>
    public override bool IsValid(object? value)
    {
        return value is Dictionary<DiscordRole, ulong> roles
            && roles.All(role => Enum.IsDefined(role.Key) && role.Value > 0);
    }
}
