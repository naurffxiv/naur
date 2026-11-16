// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration;
using Naur.Authingway.Discord;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Tests.Discord;

/// <summary>
/// Contains tests that verify the role and channel group precondition.
/// </summary>
public class RequireDiscordRolesAttributeTests
{
    private const ulong AdminRole = 10;
    private const ulong DeveloperRole = 20;
    private const ulong AdminChannel = 100;
    private const ulong DeveloperChannelA = 200;
    private const ulong DeveloperChannelB = 201;

    private static readonly RequireDiscordRolesAttribute<ApplicationCommandContext> s_attribute = new()
    {
        Roles = [DiscordRole.Admin, DiscordRole.Developer],
        Channels = [DiscordChannelGroup.Admin, DiscordChannelGroup.Developer]
    };

    /// <summary>
    /// Verifies that a member holding any given role is permitted in any channel of any given group.
    /// </summary>
    /// <param name="channelId">The channel the command is invoked from.</param>
    /// <param name="roleId">The role the member holds.</param>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    [Arguments(AdminChannel, AdminRole)]
    [Arguments(DeveloperChannelA, DeveloperRole)]
    [Arguments(DeveloperChannelB, AdminRole)]
    public async Task PermitsAnyRoleInAnyChannelGroup(ulong channelId, ulong roleId)
    {
        await Assert.That(s_attribute.IsPermitted(CreateConfig(), channelId, [roleId]))
            .IsTrue();
    }

    /// <summary>
    /// Verifies that a channel outside every given group is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesOtherChannels()
    {
        await Assert.That(s_attribute.IsPermitted(CreateConfig(), 999, [AdminRole]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that a member holding none of the given roles is refused.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesMembersWithoutAGivenRole()
    {
        await Assert.That(s_attribute.IsPermitted(CreateConfig(), AdminChannel, [999]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that an invocation from outside a guild is refused when roles are required.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesInvocationsOutsideAGuild()
    {
        await Assert.That(s_attribute.IsPermitted(CreateConfig(), AdminChannel, null))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that roles and groups with no configured identifiers match nothing, keeping the command closed.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesWhenNothingIsConfigured()
    {
        await Assert.That(s_attribute.IsPermitted(new DiscordConfig(), AdminChannel, [AdminRole]))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that leaving a property empty does not restrict that dimension.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task EmptyPropertiesDoNotRestrict()
    {
        var rolesOnly = new RequireDiscordRolesAttribute<ApplicationCommandContext> { Roles = [DiscordRole.Admin] };
        var channelsOnly = new RequireDiscordRolesAttribute<ApplicationCommandContext> { Channels = [DiscordChannelGroup.Developer] };

        await Assert.That(rolesOnly.IsPermitted(CreateConfig(), 999, [AdminRole])).IsTrue();
        await Assert.That(channelsOnly.IsPermitted(CreateConfig(), DeveloperChannelA, null)).IsTrue();
    }

    /// <summary>
    /// Verifies that configuration keys bind to the enum names without regard to case.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task BindsRolesAndChannelGroupsFromConfiguration()
    {
        var builder = Host.CreateApplicationBuilder();
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Discord:Roles:admin"] = $"{AdminRole}",
            ["Discord:Channels:admin:0"] = $"{AdminChannel}",
            ["Discord:Channels:Developer:0"] = $"{DeveloperChannelA}",
            ["Discord:Channels:Developer:1"] = $"{DeveloperChannelB}"
        });
        builder.ConfigureConfiguration();

        using var services = builder.Services.BuildServiceProvider();
        var config = services.GetRequiredService<IOptions<DiscordConfig>>().Value;

        await Assert.That(config.Roles[DiscordRole.Admin]).IsEqualTo(AdminRole);
        List<ulong> developerChannels = [DeveloperChannelA, DeveloperChannelB];

        await Assert.That(config.Channels[DiscordChannelGroup.Developer]).IsEquivalentTo(developerChannels);
        await Assert.That(s_attribute.IsPermitted(config, AdminChannel, [AdminRole])).IsTrue();
    }

    private static DiscordConfig CreateConfig()
    {
        return new DiscordConfig
        {
            Roles =
            {
                [DiscordRole.Admin] = AdminRole,
                [DiscordRole.Developer] = DeveloperRole
            },
            Channels =
            {
                [DiscordChannelGroup.Admin] = [AdminChannel],
                [DiscordChannelGroup.Developer] = [DeveloperChannelA, DeveloperChannelB]
            }
        };
    }
}
