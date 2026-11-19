// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Identity.Models;
using Naur.Authingway.Identity.Services;
using Naur.Authingway.Tests.Testing;
using OpenIddict.Abstractions;
using static OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Tests.Identity.Services;

/// <summary>
/// Tests client management against the real OpenIddict stores.
/// </summary>
/// <param name="clients">The client registration service.</param>
/// <param name="manager">The OpenIddict application manager.</param>
[AuthingwayDataSource]
public class ClientRegistrationServiceTests(ClientRegistrationService clients, IOpenIddictApplicationManager manager)
{
    /// <summary>
    /// Verifies the full lifecycle of a client, including that its secret survives updates.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ManagesClientLifecycle()
    {
        var created = await clients.CreateAsync(Settings("Lifecycle", "https://app.example/cb", Scopes.OpenId));

        await Assert.That(created.ClientId).IsNotEmpty();
        await Assert.That(created.ClientSecret.Length).IsGreaterThanOrEqualTo(43);
        await Assert.That(await ValidateSecretAsync(created)).IsTrue();

        var found = await clients.FindAsync(created.ClientId);

        await Assert.That(found).IsNotNull();
        await Assert.That(found!.ClientType).IsEqualTo(ClientTypes.Confidential);
        await Assert.That(found.Settings.DisplayName).IsEqualTo("Lifecycle");
        await Assert.That(found.Settings.RequirePkce).IsTrue();

        var updated = await clients.UpdateAsync(created.ClientId, Settings("Renamed", "https://other.example/cb", Scopes.Roles));
        var afterUpdate = await clients.FindAsync(created.ClientId);

        await Assert.That(updated).IsTrue();
        await Assert.That(afterUpdate!.Settings.DisplayName).IsEqualTo("Renamed");
        await Assert.That(afterUpdate.Settings.RedirectUris.Single()).IsEqualTo(new Uri("https://other.example/cb"));
        await Assert.That(afterUpdate.Settings.Scopes.Single()).IsEqualTo(Scopes.Roles);
        await Assert.That(await ValidateSecretAsync(created)).IsTrue();

        var listed = await ListAllAsync();

        await Assert.That(listed.Any(c => c.ClientId == created.ClientId && c.DisplayName == "Renamed")).IsTrue();

        await Assert.That(await clients.DeleteAsync(created.ClientId)).IsTrue();
        await Assert.That(await clients.FindAsync(created.ClientId)).IsNull();
    }

    /// <summary>
    /// Verifies that operations on unknown clients report that nothing was found.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ReportsMissingClients()
    {
        var clientId = Guid.NewGuid().ToString("N");

        await Assert.That(await clients.FindAsync(clientId)).IsNull();
        await Assert.That(await clients.UpdateAsync(clientId, Settings("Missing", "https://app.example/cb", Scopes.OpenId))).IsFalse();
        await Assert.That(await clients.DeleteAsync(clientId)).IsFalse();
    }

    private static ClientSettings Settings(string name, string redirectUri, string scope)
    {
        return ClientSettings.Parse(name, redirectUri, [GrantTypes.AuthorizationCode], [scope], requirePkce: true).Settings!;
    }

    private async Task<bool> ValidateSecretAsync(ClientCredentials credentials)
    {
        var application = await manager.FindByClientIdAsync(credentials.ClientId);

        return application is not null && await manager.ValidateClientSecretAsync(application, credentials.ClientSecret);
    }

    private async Task<List<ClientSummary>> ListAllAsync()
    {
        var all = new List<ClientSummary>();
        var page = 1;
        ClientPage result;

        do
        {
            result = await clients.ListAsync(page++);
            all.AddRange(result.Clients);
        }
        while (result.Page < result.PageCount);

        return all;
    }
}
