// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Text;
using NetCord;
using NetCord.Rest;

namespace Naur.Authingway.Discord.Responses;

/// <summary>
/// Builds interaction responses that only the invoking user can see.
/// </summary>
public static class EphemeralResponse
{
    private const string MarkdownCharacters = "\\*_~`|>#-[]()";

    /// <summary>
    /// Creates an ephemeral message callback that never pings anyone.
    /// </summary>
    /// <param name="content">The message content.</param>
    /// <returns>The interaction callback.</returns>
    public static InteractionCallbackProperties Message(string content)
    {
        return InteractionCallback.Message(new InteractionMessageProperties
        {
            Content = content,
            Flags = MessageFlags.Ephemeral,
            AllowedMentions = AllowedMentionsProperties.None
        });
    }

    /// <summary>
    /// Escapes Discord markdown in user-provided text.
    /// </summary>
    /// <param name="text">The text to escape.</param>
    /// <returns>The escaped text.</returns>
    public static string Escape(string? text)
    {
        var builder = new StringBuilder();

        foreach (var character in text ?? string.Empty)
        {
            if (MarkdownCharacters.Contains(character, StringComparison.Ordinal))
            {
                builder.Append('\\');
            }

            builder.Append(character);
        }

        return builder.ToString();
    }
}
