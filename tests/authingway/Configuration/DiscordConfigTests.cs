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
    /// Verifies that invalid bound Discord settings are rejected during startup.
    /// </summary>
    /// <param name="key">The configuration key holding a zero identifier.</param>
    /// <param name="property">The property expected in the validation error.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("Discord:Roles:Admin", "Roles")]
    [Arguments("Discord:Channels:Developer:0", "Channels")]
    [Arguments("Discord:ClientId", "ClientId")]
    [Arguments("Discord:ClientSecret", "ClientSecret")]
    public async Task RejectsInvalidConfigurationOnStartup(string key, string property)
    {
        var builder = Host.CreateApplicationBuilder(new HostApplicationBuilderSettings { DisableDefaults = true });
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Discord:GuildId"] = "1",
            ["Discord:ClientId"] = "123456789",
            ["Discord:ClientSecret"] = "01234567890123456789012345678901",
            [key] = "0"
        });
        builder.ConfigureConfiguration();

        using var host = builder.Build();

        await Assert.That(() => host.StartAsync())
            .Throws<OptionsValidationException>()
            .WithMessageContaining(property);
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
            ["Discord:ClientId"] = "123456789",
            ["Discord:ClientSecret"] = "01234567890123456789012345678901",
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
            .AddJsonFile(FindAppSettings($"appsettings.{environment}.json"))
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                ["Discord:ClientId"] = "123456789",
                ["Discord:ClientSecret"] = "01234567890123456789012345678901"
            });
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
