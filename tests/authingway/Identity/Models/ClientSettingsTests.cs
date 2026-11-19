// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Identity.Models;
using OpenIddict.Abstractions;
using static OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Tests.Identity.Models;

/// <summary>
/// Tests validation and descriptor mapping of client settings.
/// </summary>
public class ClientSettingsTests
{
    /// <summary>
    /// Verifies that valid form input is normalized into settings.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ParsesValidInput()
    {
        var result = ClientSettings.Parse(
            "  My App  ",
            "https://app.example/callback\r\n\r\nhttp://localhost:5000/cb\nhttp://127.0.0.1:5000/cb\nhttps://app.example/callback",
            [GrantTypes.AuthorizationCode, GrantTypes.RefreshToken],
            [Scopes.OpenId, Scopes.Roles],
            requirePkce: false);

        await Assert.That(result.Errors).IsEmpty();
        await Assert.That(result.Settings).IsNotNull();
        await Assert.That(result.Settings!.DisplayName).IsEqualTo("My App");
        await Assert.That(result.Settings.RedirectUris.Count).IsEqualTo(3);
        await Assert.That(result.Settings.GrantTypes.Count).IsEqualTo(2);
        await Assert.That(result.Settings.Scopes.Count).IsEqualTo(2);
        await Assert.That(result.Settings.RequirePkce).IsFalse();
    }

    /// <summary>
    /// Verifies that invalid input is rejected without producing settings.
    /// </summary>
    /// <param name="displayName">The display name.</param>
    /// <param name="redirectUris">The redirect URIs.</param>
    /// <param name="grantType">The selected grant type.</param>
    /// <param name="scope">The selected scope.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("", "https://app.example/cb", GrantTypes.AuthorizationCode, Scopes.OpenId)]
    [Arguments("App", "", GrantTypes.AuthorizationCode, Scopes.OpenId)]
    [Arguments("App", "/relative", GrantTypes.AuthorizationCode, Scopes.OpenId)]
    [Arguments("App", "ftp://app.example/cb", GrantTypes.AuthorizationCode, Scopes.OpenId)]
    [Arguments("App", "http://app.example/cb", GrantTypes.AuthorizationCode, Scopes.OpenId)]
    [Arguments("App", "https://app.example/cb#frag", GrantTypes.AuthorizationCode, Scopes.OpenId)]
    [Arguments("App", "https://app.example/cb", GrantTypes.RefreshToken, Scopes.OpenId)]
    [Arguments("App", "https://app.example/cb", GrantTypes.ClientCredentials, Scopes.OpenId)]
    [Arguments("App", "https://app.example/cb", GrantTypes.AuthorizationCode, Scopes.Email)]
    public async Task RejectsInvalidInput(string displayName, string redirectUris, string grantType, string scope)
    {
        var grants = grantType == GrantTypes.AuthorizationCode
            ? new[] { grantType }
            : grantType == GrantTypes.RefreshToken ? [grantType] : [GrantTypes.AuthorizationCode, grantType];

        var result = ClientSettings.Parse(displayName, redirectUris, grants, [scope], requirePkce: true);

        await Assert.That(result.Settings).IsNull();
        await Assert.That(result.Errors).IsNotEmpty();
    }

    /// <summary>
    /// Verifies the display name and redirect URI limits.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task EnforcesLimits()
    {
        var tooLong = ClientSettings.Parse(
            new string('a', ClientSettings.MaxDisplayNameLength + 1), "https://app.example/cb", [GrantTypes.AuthorizationCode], [], true);
        var tooMany = ClientSettings.Parse(
            "App",
            string.Join('\n', Enumerable.Range(0, ClientSettings.MaxRedirectUris + 1).Select(i => $"https://app.example/{i}")),
            [GrantTypes.AuthorizationCode],
            [],
            true);

        await Assert.That(tooLong.Settings).IsNull();
        await Assert.That(tooMany.Settings).IsNull();
    }

    /// <summary>
    /// Verifies that applying settings replaces only managed permissions and round-trips through a descriptor.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task AppliesAndReadsManagedValues()
    {
        var descriptor = new OpenIddictApplicationDescriptor
        {
            DisplayName = "Old",
            RedirectUris = { new Uri("https://old.example/cb") },
            Permissions = { Permissions.Endpoints.Revocation, Permissions.Prefixes.Scope + Scopes.Roles },
            Requirements = { Requirements.Features.ProofKeyForCodeExchange }
        };

        var settings = ClientSettings.Parse(
            "New", "https://new.example/cb", [GrantTypes.AuthorizationCode, GrantTypes.RefreshToken], [Scopes.OpenId], false).Settings!;

        settings.ApplyTo(descriptor);

        await Assert.That(descriptor.DisplayName).IsEqualTo("New");
        await Assert.That(descriptor.RedirectUris.Single()).IsEqualTo(new Uri("https://new.example/cb"));
        await Assert.That(descriptor.Permissions.Contains(Permissions.Endpoints.Revocation)).IsTrue();
        await Assert.That(descriptor.Permissions.Contains(Permissions.Endpoints.Authorization)).IsTrue();
        await Assert.That(descriptor.Permissions.Contains(Permissions.Endpoints.Token)).IsTrue();
        await Assert.That(descriptor.Permissions.Contains(Permissions.ResponseTypes.Code)).IsTrue();
        await Assert.That(descriptor.Permissions.Contains(Permissions.GrantTypes.RefreshToken)).IsTrue();
        await Assert.That(descriptor.Permissions.Contains(Permissions.Prefixes.Scope + Scopes.Roles)).IsFalse();
        await Assert.That(descriptor.Requirements).IsEmpty();

        var roundTrip = ClientSettings.FromDescriptor(descriptor);

        await Assert.That(roundTrip.DisplayName).IsEqualTo(settings.DisplayName);
        await Assert.That(roundTrip.RedirectUris.SequenceEqual(settings.RedirectUris)).IsTrue();
        await Assert.That(roundTrip.GrantTypes.SetEquals(settings.GrantTypes)).IsTrue();
        await Assert.That(roundTrip.Scopes.SetEquals(settings.Scopes)).IsTrue();
        await Assert.That(roundTrip.RequirePkce).IsFalse();
    }
}
