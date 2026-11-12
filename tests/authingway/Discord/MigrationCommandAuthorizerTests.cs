// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Options;
using Naur.Authingway.Discord;

namespace Naur.Authingway.Tests.Discord;

/// <summary>
/// Contains tests that verify the migration command is restricted to one guild, one channel and one role.
/// </summary>
public class MigrationCommandAuthorizerTests
{
    private const ulong GuildId = 50;
    private const ulong ChannelId = 100;
    private const ulong RoleId = 200;

    /// <summary>
    /// Verifies that an invocation from the configured guild and channel by a holder of the configured role is permitted.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AuthorizesConfiguredGuildChannelAndRole()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, [RoleId]))
            .IsTrue();
    }

    /// <summary>
    /// Verifies that an invocation from any other guild is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesOtherGuilds()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(999, ChannelId, [RoleId]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that an invocation from outside any guild is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesInvocationsOutsideAGuild()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(null, ChannelId, [RoleId]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that an invocation from any other channel is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesOtherChannels()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(GuildId, 999, [RoleId]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that a user without the configured role is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesUsersWithoutTheRole()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, [999]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that a user holding no roles is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesUsersWithNoRoles()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, []))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that nothing is permitted while the guild is unconfigured.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesWhenGuildIsUnconfigured()
    {
        var authorizer = CreateAuthorizer(0, ChannelId, RoleId);

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, [RoleId]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that nothing is permitted while the channel is unconfigured.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesWhenChannelIsUnconfigured()
    {
        var authorizer = CreateAuthorizer(GuildId, 0, RoleId);

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, [RoleId]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that nothing is permitted while the role is unconfigured.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesWhenRoleIsUnconfigured()
    {
        var authorizer = CreateAuthorizer(GuildId, ChannelId, 0);

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, [RoleId]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that blank configured identifiers bind as unconfigured, so nothing is permitted.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesWhenIdentifiersAreBlank()
    {
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                [$"{DiscordOptions.SectionName}:{nameof(DiscordOptions.GuildId)}"] = "",
                [$"{DiscordOptions.SectionName}:{nameof(DiscordOptions.MigrationChannelId)}"] = "",
                [$"{DiscordOptions.SectionName}:{nameof(DiscordOptions.DeveloperRoleId)}"] = ""
            })
            .Build();

        var options = new DiscordOptions();
        configuration.GetSection(DiscordOptions.SectionName).Bind(options);

        var authorizer = new MigrationCommandAuthorizer(Options.Create(options));

        await Assert.That(authorizer.IsAuthorized(GuildId, ChannelId, [RoleId]))
            .IsFalse();
    }

    private static MigrationCommandAuthorizer CreateAuthorizer(ulong guildId, ulong channelId, ulong roleId)
    {
        var options = Options.Create(new DiscordOptions
        {
            GuildId = guildId,
            MigrationChannelId = channelId,
            DeveloperRoleId = roleId
        });

        return new MigrationCommandAuthorizer(options);
    }
}
