// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Discord.Responses;
using NetCord;
using NetCord.Rest;

namespace Naur.Authingway.Tests.Discord.Responses;

/// <summary>
/// Tests ephemeral responses and Markdown escaping.
/// </summary>
public class EphemeralResponseTests
{
    /// <summary>
    /// Verifies that responses are visible only to the caller and never ping anyone.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task MessagesAreEphemeralAndMentionNobody()
    {
        var callback = (InteractionCallbackProperties<InteractionMessageProperties>)EphemeralResponse.Message("hi @everyone");

        await Assert.That(callback.Data.Content).IsEqualTo("hi @everyone");
        await Assert.That(callback.Data.Flags).IsEqualTo(MessageFlags.Ephemeral);
        var mentions = callback.Data.AllowedMentions!;
        await Assert.That(mentions.Everyone).IsFalse();
        await Assert.That(mentions.ReplyMention).IsFalse();
        await Assert.That(mentions.AllowedUsers).IsEmpty();
        await Assert.That(mentions.AllowedRoles).IsEmpty();
    }

    /// <summary>
    /// Verifies that every Markdown control character is escaped.
    /// </summary>
    /// <param name="input">The raw text.</param>
    /// <param name="expected">The escaped text.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("plain", "plain")]
    [Arguments("*bold* _it_ ~s~ `c` |sp| >q #h -l [a](b)", "\\*bold\\* \\_it\\_ \\~s\\~ \\`c\\` \\|sp\\| \\>q \\#h \\-l \\[a\\]\\(b\\)")]
    [Arguments("back\\slash", "back\\\\slash")]
    [Arguments(null, "")]
    public async Task EscapesMarkdown(string? input, string expected)
    {
        await Assert.That(EphemeralResponse.Escape(input)).IsEqualTo(expected);
    }
}
