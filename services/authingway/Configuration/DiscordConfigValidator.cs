// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Globalization;
using Microsoft.Extensions.Options;

namespace Naur.Authingway.Configuration;

/// <summary>
/// Rejects <c>Discord:Roles</c> and <c>Discord:Channels</c> entries that name no <see cref="DiscordRole"/> or
/// <see cref="DiscordChannelGroup"/>, or whose identifiers are not positive numbers.
/// </summary>
/// <param name="configuration">The application configuration whose raw keys are checked. Cannot be null.</param>
/// <remarks>
/// The configuration binder drops dictionary keys it cannot convert, so the raw keys are checked here instead of the
/// bound <see cref="DiscordConfig"/>. Every problem is reported in a single failure.
/// </remarks>
public sealed class DiscordConfigValidator(IConfiguration configuration) : IValidateOptions<DiscordConfig>
{
    /// <inheritdoc/>
    public ValidateOptionsResult Validate(string? name, DiscordConfig options)
    {
        var failures = new List<string>();

        foreach (var role in configuration.GetSection("Discord:Roles").GetChildren())
        {
            if (!Enum.TryParse<DiscordRole>(role.Key, ignoreCase: true, out _))
            {
                failures.Add($"Discord:Roles:{role.Key} is not a known role. Known roles: {string.Join(", ", Enum.GetNames<DiscordRole>())}.");
            }

            if (!IsIdentifier(role.Value))
            {
                failures.Add($"Discord:Roles:{role.Key} must be a positive role identifier, but was '{role.Value}'.");
            }
        }

        foreach (var group in configuration.GetSection("Discord:Channels").GetChildren())
        {
            if (!Enum.TryParse<DiscordChannelGroup>(group.Key, ignoreCase: true, out _))
            {
                failures.Add($"Discord:Channels:{group.Key} is not a known channel group. Known groups: {string.Join(", ", Enum.GetNames<DiscordChannelGroup>())}.");
            }

            foreach (var channel in group.GetChildren())
            {
                if (!IsIdentifier(channel.Value))
                {
                    failures.Add($"Discord:Channels:{group.Key}:{channel.Key} must be a positive channel identifier, but was '{channel.Value}'.");
                }
            }
        }

        return failures.Count == 0 ? ValidateOptionsResult.Success : ValidateOptionsResult.Fail(failures);
    }

    private static bool IsIdentifier(string? value)
    {
        return ulong.TryParse(value, NumberStyles.None, CultureInfo.InvariantCulture, out var id) && id != 0;
    }
}
