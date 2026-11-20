// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Text;
using Naur.Authingway.Discord.Responses;
using Naur.Authingway.Identity.Models;

namespace Naur.Authingway.Identity.Responses;

/// <summary>
/// Formats client information for Discord messages. Secrets are only included when newly generated.
/// </summary>
public static class ClientMessages
{
    /// <summary>
    /// Describes a client's stored settings.
    /// </summary>
    /// <param name="client">The client to describe.</param>
    /// <returns>The formatted description.</returns>
    public static string Describe(ClientDetails client)
    {
        ArgumentNullException.ThrowIfNull(client);

        var settings = client.Settings;
        var builder = new StringBuilder()
            .AppendLine($"**{EphemeralResponse.Escape(settings.DisplayName)}**")
            .AppendLine($"Client ID: `{client.ClientId}`")
            .AppendLine($"Client type: {client.ClientType ?? "unknown"}")
            .AppendLine("Redirect URIs:");

        foreach (var uri in settings.RedirectUris)
        {
            builder.AppendLine($"- <{uri.AbsoluteUri}>");
        }

        return builder
            .AppendLine($"Grant types: {Join(ClientSettings.SupportedGrantTypes.Where(settings.GrantTypes.Contains))}")
            .AppendLine($"Scopes: {Join(ClientSettings.SupportedScopes.Where(settings.Scopes.Contains))}")
            .Append($"PKCE required: {(settings.RequirePkce ? "yes" : "no")}")
            .ToString();
    }

    /// <summary>
    /// Formats validation errors.
    /// </summary>
    /// <param name="errors">The validation errors.</param>
    /// <returns>The formatted errors.</returns>
    public static string Invalid(IEnumerable<string> errors)
    {
        return "The client was not saved:\n" + string.Join('\n', errors.Select(error => $"- {error}"));
    }

    /// <summary>
    /// Formats a newly generated client secret, which is shown only once.
    /// </summary>
    /// <param name="clientSecret">The plain-text client secret.</param>
    /// <returns>The formatted message.</returns>
    public static string Secret(string clientSecret)
    {
        return $"Client secret: `{clientSecret}`\n" +
            "Copy the secret now. It is stored hashed and cannot be shown again.";
    }

    /// <summary>
    /// Formats the message for a client identifier that does not exist.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <returns>The formatted message.</returns>
    public static string NotFound(string clientId)
    {
        return $"No client with ID `{clientId.Replace('`', '\'')}` exists.";
    }

    private static string Join(IEnumerable<string> values)
    {
        var list = values.Select(value => $"`{value}`").ToList();

        return list.Count > 0 ? string.Join(", ", list) : "none";
    }
}
