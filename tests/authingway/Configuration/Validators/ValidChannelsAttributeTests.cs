// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Configuration.Validators;

namespace Naur.Authingway.Tests.Configuration.Validators;

/// <summary>
/// Tests validation of Discord channel dictionaries.
/// </summary>
public class ValidChannelsAttributeTests
{
    /// <summary>
    /// Verifies that empty dictionaries, empty groups and positive identifiers are accepted.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task AcceptsValidChannels()
    {
        var attribute = new ValidChannelsAttribute();

        await Assert.That(attribute.IsValid(new Dictionary<DiscordChannelGroup, HashSet<ulong>>())).IsTrue();
        await Assert.That(attribute.IsValid(new Dictionary<DiscordChannelGroup, HashSet<ulong>>
        {
            [DiscordChannelGroup.Admin] = [],
            [DiscordChannelGroup.Developer] = [1, ulong.MaxValue]
        })).IsTrue();
    }

    /// <summary>
    /// Verifies that undefined groups and zero identifiers are rejected alongside valid entries.
    /// </summary>
    /// <param name="group">The channel group to validate.</param>
    /// <param name="identifier">The channel identifier to validate.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments((DiscordChannelGroup)99, 10UL)]
    [Arguments(DiscordChannelGroup.Developer, 0UL)]
    public async Task RejectsInvalidChannels(DiscordChannelGroup group, ulong identifier)
    {
        var channels = new Dictionary<DiscordChannelGroup, HashSet<ulong>>
        {
            [DiscordChannelGroup.Admin] = [1],
            [group] = [2, identifier]
        };

        await Assert.That(new ValidChannelsAttribute().IsValid(channels)).IsFalse();
    }

    /// <summary>
    /// Verifies that null collections, null groups and incorrectly typed values are rejected.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task RejectsMissingOrIncorrectlyTypedValues()
    {
        var attribute = new ValidChannelsAttribute();

        await Assert.That(attribute.IsValid(null)).IsFalse();
        await Assert.That(attribute.IsValid("Admin")).IsFalse();
        await Assert.That(attribute.IsValid(new Dictionary<DiscordChannelGroup, HashSet<ulong>>
        {
            [DiscordChannelGroup.Admin] = null!
        })).IsFalse();
    }
}
