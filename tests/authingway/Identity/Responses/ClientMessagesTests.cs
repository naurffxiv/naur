// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Identity.Models;
using Naur.Authingway.Identity.Responses;
using static OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Tests.Identity.Responses;

/// <summary>
/// Tests the text shown to Discord users about clients.
/// </summary>
public class ClientMessagesTests
{
    /// <summary>
    /// Verifies that a client description lists its settings in display order and escapes the display name.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task DescribesClientSettings()
    {
        var settings = ClientSettings.Parse(
            "My *App*", "https://app.example/cb\nhttp://localhost:5000/cb",
            [GrantTypes.RefreshToken, GrantTypes.AuthorizationCode], [Scopes.OpenId], requirePkce: true).Settings!;

        var text = ClientMessages.Describe(new ClientDetails("abc", ClientTypes.Confidential, settings));

        await Assert.That(text).IsEqualTo(
            "**My \\*App\\***\n" +
            "Client ID: `abc`\n" +
            "Client type: confidential\n" +
            "Redirect URIs:\n" +
            "- <https://app.example/cb>\n" +
            "- <http://localhost:5000/cb>\n" +
            "Grant types: `authorization_code`, `refresh_token`\n" +
            "Scopes: `openid`\n" +
            "PKCE required: yes");
    }

    /// <summary>
    /// Verifies the description when no scopes are granted and the client type is unknown.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task DescribesMissingValues()
    {
        var settings = ClientSettings.Parse("App", "https://app.example/cb", [GrantTypes.AuthorizationCode], [], requirePkce: false).Settings!;

        var text = ClientMessages.Describe(new ClientDetails("abc", null, settings));

        await Assert.That(text).Contains("Client type: unknown\n");
        await Assert.That(text).Contains("Scopes: none\n");
        await Assert.That(text).EndsWith("PKCE required: no");
    }

    /// <summary>
    /// Verifies the validation, secret, and not-found messages.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task FormatsOtherMessages()
    {
        await Assert.That(ClientMessages.Invalid(["First.", "Second."])).IsEqualTo("The client was not saved:\n- First.\n- Second.");
        await Assert.That(ClientMessages.Secret("s3cret")).IsEqualTo(
            "Client secret: `s3cret`\nCopy the secret now. It is stored hashed and cannot be shown again.");
        await Assert.That(ClientMessages.NotFound("a`b")).IsEqualTo("No client with ID `a'b` exists.");
    }
}
