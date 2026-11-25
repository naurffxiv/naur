// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Preconditions;
using NetCord;
using NetCord.Gateway;
using NetCord.JsonModels;
using NetCord.Services;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Tests.Discord.Preconditions;

/// <summary>
/// Tests Discord authorization through real application-command contexts and configured options.
/// </summary>
public class AuthorizeAttributesTests
{
    /// <summary>
    /// Verifies that only the configured guild is accepted, even when the guild is not cached.
    /// </summary>
    /// <param name="guildId">The invoking guild, or null for a direct message.</param>
    /// <param name="expected">Whether access should be granted.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments(1UL, true)]
    [Arguments(2UL, false)]
    [Arguments(0UL, false)]
    [Arguments(null, false)]
    public async Task AuthorizesConfiguredGuild(ulong? guildId, bool expected)
    {
        await Assert.That(await IsPermittedAsync(new AuthorizeGuildAttribute(), CreateConfig(), guildId: guildId))
            .IsEqualTo(expected);
    }

    /// <summary>
    /// Verifies that an unset guild identifier never grants access.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task RejectsUnconfiguredGuild()
    {
        await Assert.That(await IsPermittedAsync(new AuthorizeGuildAttribute(), new DiscordConfig(), guildId: 0))
            .IsFalse();
    }

    /// <summary>
    /// Verifies any-of role matching, including members without roles and direct messages.
    /// </summary>
    /// <param name="roleIds">The invoking member's roles.</param>
    /// <param name="guildId">The invoking guild.</param>
    /// <param name="expected">Whether access should be granted.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments(new ulong[] { 10 }, 1UL, true)]
    [Arguments(new ulong[] { 999, 20 }, 1UL, true)]
    [Arguments(new ulong[] { 999 }, 1UL, false)]
    [Arguments(new ulong[] { }, 1UL, false)]
    [Arguments(new ulong[] { 10 }, null, false)]
    public async Task AuthorizesAnyRequestedRole(ulong[] roleIds, ulong? guildId, bool expected)
    {
        var attribute = new AuthorizeRolesAttribute([DiscordRole.Admin, DiscordRole.Developer]);

        await Assert.That(await IsPermittedAsync(attribute, CreateConfig(), guildId: guildId, roleIds: roleIds))
            .IsEqualTo(expected);
    }

    /// <summary>
    /// Verifies that any channel in any requested group is accepted.
    /// </summary>
    /// <param name="channelId">The invoking channel.</param>
    /// <param name="expected">Whether access should be granted.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments(100UL, true)]
    [Arguments(200UL, true)]
    [Arguments(201UL, true)]
    [Arguments(999UL, false)]
    [Arguments(0UL, false)]
    public async Task AuthorizesAnyRequestedChannelGroup(ulong channelId, bool expected)
    {
        var attribute = new AuthorizeChannelsAttribute([DiscordChannelGroup.Admin, DiscordChannelGroup.Developer]);

        await Assert.That(await IsPermittedAsync(attribute, CreateConfig(), channelId: channelId))
            .IsEqualTo(expected);
    }

    /// <summary>
    /// Verifies empty lists, missing mappings, empty groups, and zero identifiers fail closed.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task RejectsEmptyOrUnconfiguredAllowLists()
    {
        var config = CreateConfig();

        await Assert.That(await IsPermittedAsync(new AuthorizeRolesAttribute([]), config)).IsFalse();
        await Assert.That(await IsPermittedAsync(new AuthorizeChannelsAttribute([]), config)).IsFalse();

        var roles = new AuthorizeRolesAttribute([DiscordRole.Admin]);
        var channels = new AuthorizeChannelsAttribute([DiscordChannelGroup.Admin]);
        var missing = new DiscordConfig();

        await Assert.That(await IsPermittedAsync(roles, missing)).IsFalse();
        await Assert.That(await IsPermittedAsync(channels, missing)).IsFalse();

        config.Roles[DiscordRole.Admin] = 0;
        config.Channels[DiscordChannelGroup.Admin] = [];

        await Assert.That(await IsPermittedAsync(roles, config, roleIds: [0])).IsFalse();
        await Assert.That(await IsPermittedAsync(channels, config)).IsFalse();

        config.Channels[DiscordChannelGroup.Admin] = [0];

        await Assert.That(await IsPermittedAsync(channels, config, channelId: 0)).IsFalse();
    }

    /// <summary>
    /// Verifies missing mappings cannot mask a valid match or grant access through unrequested mappings.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task UsesOnlyRequestedConfiguredMappings()
    {
        var config = CreateConfig();
        var adminOnly = new AuthorizeRolesAttribute([DiscordRole.Admin]);
        var adminChannelsOnly = new AuthorizeChannelsAttribute([DiscordChannelGroup.Admin]);

        await Assert.That(await IsPermittedAsync(adminOnly, config, roleIds: [20])).IsFalse();
        await Assert.That(await IsPermittedAsync(adminChannelsOnly, config, channelId: 200)).IsFalse();

        config.Roles.Remove(DiscordRole.Admin);
        config.Channels.Remove(DiscordChannelGroup.Admin);

        await Assert.That(await IsPermittedAsync(
            new AuthorizeRolesAttribute([DiscordRole.Admin, DiscordRole.Developer]), config, roleIds: [20])).IsTrue();
        await Assert.That(await IsPermittedAsync(
            new AuthorizeChannelsAttribute([DiscordChannelGroup.Admin, DiscordChannelGroup.Developer]), config, channelId: 201)).IsTrue();
    }

    /// <summary>
    /// Verifies each attribute restricts only its own dimension.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task AppliesRestrictionsIndependently()
    {
        var config = CreateConfig();

        await Assert.That(await IsPermittedAsync(new AuthorizeGuildAttribute(), config, channelId: 999, roleIds: [])).IsTrue();
        await Assert.That(await IsPermittedAsync(
            new AuthorizeRolesAttribute([DiscordRole.Admin]), config, guildId: 2, channelId: 999)).IsTrue();
        await Assert.That(await IsPermittedAsync(
            new AuthorizeChannelsAttribute([DiscordChannelGroup.Admin]), config, guildId: null, roleIds: [])).IsTrue();
    }

    /// <summary>
    /// Verifies NetCord discovers the non-generic attributes and requires all stacked checks to succeed.
    /// </summary>
    /// <param name="guildId">The invoking guild.</param>
    /// <param name="channelId">The invoking channel.</param>
    /// <param name="roleId">The invoking member's role.</param>
    /// <param name="expected">Whether execution should succeed.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments(1UL, 100UL, 10UL, true)]
    [Arguments(1UL, 201UL, 20UL, true)]
    [Arguments(2UL, 100UL, 10UL, false)]
    [Arguments(1UL, 999UL, 10UL, false)]
    [Arguments(1UL, 100UL, 999UL, false)]
    public async Task EnforcesStackedAttributes(ulong guildId, ulong channelId, ulong roleId, bool expected)
    {
        using var services = CreateServices(CreateConfig());
        using var client = CreateClient();
        var context = CreateContext(client, guildId, channelId, [roleId]);
        var commands = new ApplicationCommandService<ApplicationCommandContext>();
        commands.AddModule<AuthorizedModule>();

        var command = (SlashCommandInfo<ApplicationCommandContext>)commands.GetCommands().Single();
        await Assert.That(command.Preconditions.Count).IsEqualTo(3);

        var result = await commands.ExecuteAsync(context, services);

        await Assert.That(result is IFailResult).IsEqualTo(!expected);
    }

    private static async Task<bool> IsPermittedAsync(
        PreconditionAttribute<ApplicationCommandContext> attribute,
        DiscordConfig config,
        ulong? guildId = 1,
        ulong channelId = 100,
        ulong[]? roleIds = null)
    {
        using var services = CreateServices(config);
        using var client = CreateClient();
        var result = await attribute.EnsureCanExecuteAsync(CreateContext(client, guildId, channelId, roleIds ?? [10]), services);

        return result is PreconditionSuccessResult;
    }

    private static ServiceProvider CreateServices(DiscordConfig config)
    {
        return new ServiceCollection()
            .AddSingleton(Options.Create(config))
            .BuildServiceProvider();
    }

    private static GatewayClient CreateClient()
    {
        return new GatewayClient(new BotToken("MTIzNDU2Nzg5MDEyMzQ1Njc4.test.test"));
    }

    private static ApplicationCommandContext CreateContext(GatewayClient client, ulong? guildId, ulong channelId, ulong[] roleIds)
    {
        var user = new JsonUser { Id = 42, Username = "tester" };
        var interaction = Interaction.CreateFromJson(new JsonInteraction
        {
            Type = InteractionType.ApplicationCommand,
            GuildId = guildId,
            Channel = new JsonChannel { Id = channelId, Type = ChannelType.TextGuildChannel },
            User = user,
            GuildUser = new JsonGuildUser { User = user, RoleIds = roleIds },
            Data = new JsonInteractionData { Name = "authorized", Type = ApplicationCommandType.ChatInput },
            Entitlements = []
        }, client.Cache, client.Rest);

        return new ApplicationCommandContext((ApplicationCommandInteraction)interaction, client);
    }

    private static DiscordConfig CreateConfig()
    {
        return new DiscordConfig
        {
            GuildId = 1,
            Roles =
            {
                [DiscordRole.Admin] = 10,
                [DiscordRole.Developer] = 20
            },
            Channels =
            {
                [DiscordChannelGroup.Admin] = [100],
                [DiscordChannelGroup.Developer] = [200, 201]
            }
        };
    }

    /// <summary>
    /// Provides a command exercising the public attribute syntax through NetCord registration and execution.
    /// </summary>
    [AuthorizeGuild]
    public class AuthorizedModule : ApplicationCommandModule<ApplicationCommandContext>
    {
        /// <summary>
        /// Completes without making any Discord API calls.
        /// </summary>
        [SlashCommand("authorized", "Tests authorization.")]
        [AuthorizeRoles([DiscordRole.Admin, DiscordRole.Developer])]
        [AuthorizeChannels([DiscordChannelGroup.Admin, DiscordChannelGroup.Developer])]
        public static void Execute()
        {
        }
    }
}
