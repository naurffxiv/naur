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
        await Assert.That(roundTrip.IsPublic).IsFalse();
        await Assert.That(descriptor.ClientType).IsEqualTo(ClientTypes.Confidential);
    }

    /// <summary>
    /// Verifies that public clients always require PKCE and round-trip their client type.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task PublicClientsRequirePkce()
    {
        var settings = ClientSettings.Parse(
            "App", "https://app.example/cb", [GrantTypes.AuthorizationCode], [Scopes.OpenId], requirePkce: false, isPublic: true).Settings!;

        await Assert.That(settings.IsPublic).IsTrue();
        await Assert.That(settings.RequirePkce).IsTrue();

        var descriptor = new OpenIddictApplicationDescriptor();
        settings.ApplyTo(descriptor);

        await Assert.That(descriptor.ClientType).IsEqualTo(ClientTypes.Public);
        await Assert.That(descriptor.Requirements.Contains(Requirements.Features.ProofKeyForCodeExchange)).IsTrue();
        await Assert.That(ClientSettings.FromDescriptor(descriptor).IsPublic).IsTrue();
        await Assert.That(ClientSettings.FromDescriptor(descriptor).RequirePkce).IsTrue();
    }

    /// <summary>
    /// Verifies that new clients default to a confidential authorization code client with PKCE.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task DefaultsToConfidentialAuthorizationCodeClientWithPkce()
    {
        var settings = ClientSettings.Default;

        await Assert.That(settings.DisplayName).IsEqualTo(string.Empty);
        await Assert.That(settings.RedirectUris).IsEmpty();
        await Assert.That(settings.GrantTypes.SetEquals([GrantTypes.AuthorizationCode])).IsTrue();
        await Assert.That(settings.Scopes.SetEquals([Scopes.OpenId])).IsTrue();
        await Assert.That(settings.RequirePkce).IsTrue();
        await Assert.That(settings.IsPublic).IsFalse();
    }

    /// <summary>
    /// Verifies that input exactly at the display name and redirect URI limits is accepted.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task AcceptsInputAtTheLimits()
    {
        var result = ClientSettings.Parse(
            new string('a', ClientSettings.MaxDisplayNameLength),
            RedirectUris(ClientSettings.MaxRedirectUris),
            [GrantTypes.AuthorizationCode],
            [],
            true);

        await Assert.That(result.Errors).IsEmpty();
        await Assert.That(result.Settings!.RedirectUris.Count).IsEqualTo(ClientSettings.MaxRedirectUris);
    }

    /// <summary>
    /// Verifies that each kind of invalid input reports its own error, and only that error.
    /// </summary>
    /// <param name="displayName">The display name.</param>
    /// <param name="redirectUris">The redirect URIs.</param>
    /// <param name="grants">The selected grant types.</param>
    /// <param name="expected">The expected error.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("", "https://app.example/cb", new[] { GrantTypes.AuthorizationCode }, "A display name is required.")]
    [Arguments("App", "", new[] { GrantTypes.AuthorizationCode }, "At least one redirect URI is required.")]
    [Arguments("App", "/relative", new[] { GrantTypes.AuthorizationCode }, "`/relative` is not an absolute HTTPS URI, or an HTTP URI for localhost or 127.0.0.1, without a fragment.")]
    [Arguments("App", "https://app.example/cb", new[] { GrantTypes.RefreshToken }, "The authorization code grant is required.")]
    [Arguments("App", "https://app.example/cb", new[] { GrantTypes.AuthorizationCode, GrantTypes.ClientCredentials }, "Only supported grant types and scopes can be selected.")]
    public async Task ReportsTheSpecificError(string displayName, string redirectUris, string[] grants, string expected)
    {
        var result = ClientSettings.Parse(displayName, redirectUris, grants, [Scopes.OpenId], requirePkce: true);

        await Assert.That(result.Errors.SequenceEqual([expected])).IsTrue();
    }

    /// <summary>
    /// Verifies the errors reported for input over the limits.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ReportsLimitErrors()
    {
        var result = ClientSettings.Parse(
            new string('a', ClientSettings.MaxDisplayNameLength + 1),
            RedirectUris(ClientSettings.MaxRedirectUris + 1),
            [GrantTypes.AuthorizationCode],
            [],
            true);

        await Assert.That(result.Errors.SequenceEqual(
        [
            $"The display name must be at most {ClientSettings.MaxDisplayNameLength} characters.",
            $"At most {ClientSettings.MaxRedirectUris} redirect URIs are allowed."
        ])).IsTrue();
    }

    /// <summary>
    /// Verifies that descriptor mapping rejects a null descriptor.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task RejectsNullDescriptor()
    {
        await Assert.That(() => ClientSettings.FromDescriptor(null!)).Throws<ArgumentNullException>();
        await Assert.That(() => ClientSettings.Default.ApplyTo(null!)).Throws<ArgumentNullException>();
    }

    private static string RedirectUris(int count)
    {
        return string.Join('\n', Enumerable.Range(0, count).Select(i => $"https://app.example/{i}"));
    }
}
