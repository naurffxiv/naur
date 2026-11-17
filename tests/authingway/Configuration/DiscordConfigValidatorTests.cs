// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration;

namespace Naur.Authingway.Tests.Configuration;

/// <summary>
/// Contains tests that verify the Discord configuration is rejected when it names unknown roles or channel groups or
/// holds invalid identifiers.
/// </summary>
public class DiscordConfigValidatorTests
{
    /// <summary>
    /// Verifies that known names with positive identifiers pass validation.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AcceptsKnownNamesWithIdentifiers()
    {
        var result = Validate(new()
        {
            ["Discord:Roles:Admin"] = "10",
            ["Discord:Roles:developer"] = "20",
            ["Discord:Channels:Admin:0"] = "100",
            ["Discord:Channels:Developer:0"] = "200",
            ["Discord:Channels:Developer:1"] = "201"
        });

        await Assert.That(result.Succeeded).IsTrue();
    }

    /// <summary>
    /// Verifies that role and channel group names matching no enum value are reported with the known names.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReportsUnknownNames()
    {
        var result = Validate(new()
        {
            ["Discord:Roles:Admn"] = "10",
            ["Discord:Channels:Staff:0"] = "100"
        });

        await Assert.That(result.Failed).IsTrue();
        await Assert.That(result.FailureMessage).Contains("Discord:Roles:Admn is not a known role. Known roles: Admin, Developer.");
        await Assert.That(result.FailureMessage).Contains("Discord:Channels:Staff is not a known channel group. Known groups: Admin, Developer.");
    }

    /// <summary>
    /// Verifies that blank, zero and non-numeric identifiers are each reported.
    /// </summary>
    /// <param name="value">The identifier value to configure.</param>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    [Arguments("")]
    [Arguments("0")]
    [Arguments("abc")]
    [Arguments("-5")]
    public async Task ReportsInvalidIdentifiers(string value)
    {
        var result = Validate(new()
        {
            ["Discord:Roles:Admin"] = value,
            ["Discord:Channels:Developer:0"] = value
        });

        await Assert.That(result.FailureMessage).Contains($"Discord:Roles:Admin must be a positive role identifier, but was '{value}'.");
        await Assert.That(result.FailureMessage).Contains($"Discord:Channels:Developer:0 must be a positive channel identifier, but was '{value}'.");
    }

    /// <summary>
    /// Verifies that the host refuses to start when the Discord configuration is invalid.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task FailsStartupOnInvalidConfiguration()
    {
        var builder = Host.CreateApplicationBuilder();
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Discord:Roles:Admn"] = "10"
        });
        builder.ConfigureConfiguration();

        using var host = builder.Build();

        await Assert.That(() => host.StartAsync())
            .Throws<OptionsValidationException>()
            .WithMessageContaining("Discord:Roles:Admn");
    }

    /// <summary>
    /// Verifies that the committed <c>appsettings.json</c> of authingway passes validation.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task CommittedAppSettingsAreValid()
    {
        var configuration = new ConfigurationBuilder()
            .AddJsonFile(FindAppSettings())
            .Build();

        var result = new DiscordConfigValidator(configuration).Validate(null, new DiscordConfig());

        await Assert.That(result.FailureMessage).IsNull();
        await Assert.That(configuration.GetSection("Discord:Roles").GetChildren()).IsNotEmpty();
    }

    private static ValidateOptionsResult Validate(Dictionary<string, string?> values)
    {
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(values)
            .Build();

        return new DiscordConfigValidator(configuration).Validate(null, new DiscordConfig());
    }

    private static string FindAppSettings()
    {
        for (var directory = new DirectoryInfo(AppContext.BaseDirectory); directory is not null; directory = directory.Parent)
        {
            var path = Path.Combine(directory.FullName, "services", "authingway", "appsettings.json");

            if (File.Exists(path))
            {
                return path;
            }
        }

        throw new FileNotFoundException("Could not find services/authingway/appsettings.json above the test output directory.");
    }
}
