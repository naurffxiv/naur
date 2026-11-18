// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration;
using Naur.Authingway.Configuration.Discord;

namespace Naur.Authingway.Tests.Configuration;

/// <summary>
/// Tests Discord validation through the configured options pipeline.
/// </summary>
public class DiscordConfigTests
{
    /// <summary>
    /// Verifies that both attributes reject invalid bound identifiers during startup.
    /// </summary>
    /// <param name="key">The configuration key holding a zero identifier.</param>
    /// <param name="property">The property expected in the validation error.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("Discord:Roles:Admin", "Roles")]
    [Arguments("Discord:Channels:Developer:0", "Channels")]
    public async Task RejectsInvalidIdentifiersOnStartup(string key, string property)
    {
        var builder = Host.CreateApplicationBuilder(new HostApplicationBuilderSettings { DisableDefaults = true });
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Discord:GuildId"] = "1",
            [key] = "0"
        });
        builder.ConfigureConfiguration();

        using var host = builder.Build();

        await Assert.That(() => host.StartAsync())
            .Throws<OptionsValidationException>()
            .WithMessageContaining(property);
    }

    /// <summary>
    /// Verifies that binding rejects unknown names and malformed identifiers instead of dropping entries.
    /// </summary>
    /// <param name="key">The configuration key to bind.</param>
    /// <param name="value">The invalid configuration value.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("Discord:Roles:Admn", "10")]
    [Arguments("Discord:Channels:Staff:0", "100")]
    [Arguments("Discord:Roles:Admin", "")]
    [Arguments("Discord:Roles:Admin", "abc")]
    [Arguments("Discord:Roles:Admin", "-1")]
    [Arguments("Discord:Roles:Admin", "18446744073709551616")]
    [Arguments("Discord:Channels:Admin:0", "")]
    [Arguments("Discord:Channels:Admin:0", "abc")]
    [Arguments("Discord:Channels:Admin:0", "-1")]
    [Arguments("Discord:Channels:Admin:0", "18446744073709551616")]
    public async Task RejectsUnbindableConfiguration(string key, string value)
    {
        var builder = Host.CreateApplicationBuilder(new HostApplicationBuilderSettings { DisableDefaults = true });
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Discord:GuildId"] = "1",
            [key] = value
        });
        builder.ConfigureConfiguration();

        using var host = builder.Build();

        await Assert.That(() => host.StartAsync()).Throws<InvalidOperationException>();
    }

    /// <summary>
    /// Verifies that valid bound roles and channels pass data-annotation validation.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task AcceptsValidConfiguration()
    {
        var builder = Host.CreateApplicationBuilder(new HostApplicationBuilderSettings { DisableDefaults = true });
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Discord:GuildId"] = "1",
            ["Discord:Token"] = "test-token",
            ["Discord:Roles:Admin"] = "10",
            ["Discord:Roles:developer"] = "20",
            ["Discord:Channels:Admin:0"] = "100",
            ["Discord:Channels:Developer:0"] = "200",
            ["Discord:Channels:Developer:1"] = "201"
        });
        builder.ConfigureConfiguration();

        using var host = builder.Build();
        var options = host.Services.GetRequiredService<IOptions<DiscordConfig>>().Value;

        await Assert.That(options.Roles[DiscordRole.Developer]).IsEqualTo(20UL);
        await Assert.That(options.Channels[DiscordChannelGroup.Developer].Count).IsEqualTo(2);
    }

    /// <summary>
    /// Verifies that committed Discord settings bind and pass validation for each environment.
    /// </summary>
    /// <param name="environment">The environment whose settings should be loaded.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("Development")]
    [Arguments("Production")]
    public async Task CommittedAppSettingsAreValid(string environment)
    {
        var builder = Host.CreateApplicationBuilder(new HostApplicationBuilderSettings { DisableDefaults = true });
        builder.Configuration
            .AddJsonFile(FindAppSettings("appsettings.json"))
            .AddJsonFile(FindAppSettings($"appsettings.{environment}.json"));
        builder.ConfigureConfiguration();

        using var host = builder.Build();
        var options = host.Services.GetRequiredService<IOptions<DiscordConfig>>().Value;

        await Assert.That(options.Roles).IsNotEmpty();
        await Assert.That(options.Channels).IsNotEmpty();
    }

    private static string FindAppSettings(string fileName)
    {
        for (var directory = new DirectoryInfo(AppContext.BaseDirectory); directory is not null; directory = directory.Parent)
        {
            var path = Path.Combine(directory.FullName, "services", "authingway", fileName);

            if (File.Exists(path))
            {
                return path;
            }
        }

        throw new FileNotFoundException($"Could not find services/authingway/{fileName} above the test output directory.");
    }
}
