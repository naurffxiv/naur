// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Options;
using Naur.Authingway.Tests.Testing;
using OpenIddict.Abstractions;
using OpenIddict.Client;
using OpenIddict.Server;
using static OpenIddict.Client.WebIntegration.OpenIddictClientWebIntegrationConstants;

namespace Naur.Authingway.Tests.Identity;

/// <summary>
/// Tests the OpenIddict configuration used by the external Discord login flow.
/// </summary>
/// <param name="serviceProvider">The service provider for the running Authingway test application.</param>
[AuthingwayDataSource]
public class IdentityConfigurationTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that Discord is registered with the callback used by the login endpoint.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task DiscordClientUsesLoginCallback()
    {
        var options = serviceProvider.GetRequiredService<IOptionsMonitor<OpenIddictClientOptions>>().CurrentValue;
        var registration = options.Registrations.SingleOrDefault(r => r.ProviderName == Providers.Discord);

        await Assert.That(registration).IsNotNull();
        await Assert.That(registration!.RedirectUri).IsNotNull();
        await Assert.That(registration.RedirectUri!.OriginalString).IsEqualTo("callback/login/discord");
    }

    /// <summary>
    /// Verifies that role claims can be requested from the authorization server.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ServerRegistersRolesScope()
    {
        var options = serviceProvider.GetRequiredService<IOptionsMonitor<OpenIddictServerOptions>>().CurrentValue;

        await Assert.That(options.Scopes.Contains(OpenIddictConstants.Scopes.Roles)).IsTrue();
    }
}
