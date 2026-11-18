// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.ComponentModel.DataAnnotations;
using Naur.Authingway.Configuration.Discord;

namespace Naur.Authingway.Configuration.Validators;

/// <summary>
/// Validates that Discord channel groups have known names and positive channel identifiers.
/// </summary>
[AttributeUsage(AttributeTargets.Property | AttributeTargets.Field | AttributeTargets.Parameter)]
public sealed class ValidChannelsAttribute : ValidationAttribute
{
    /// <summary>
    /// Initializes a new instance of the <see cref="ValidChannelsAttribute"/> class.
    /// </summary>
    public ValidChannelsAttribute()
        : base("Channels must contain only known Discord channel groups with non-null sets of positive identifiers.")
    {
    }

    /// <inheritdoc/>
    public override bool IsValid(object? value)
    {
        return value is Dictionary<DiscordChannelGroup, HashSet<ulong>> channels
            && channels.All(group => Enum.IsDefined(group.Key)
                && group.Value is not null
                && group.Value.All(id => id > 0));
    }
}
