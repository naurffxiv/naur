// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using Naur.Authingway.Discord;
using NetCord.Gateway;

namespace Naur.Authingway.Tests.Discord;

/// <summary>
/// Contains tests that verify Discord gateway registration responds to the configured token value.
/// </summary>
/// <remarks>
/// These tests build a host application builder directly and never start it, so they require neither a container
/// runtime nor a Discord connection.
/// </remarks>
public class DiscordConfigurationTests
{
    /// <summary>
    /// Verifies that a gateway client is registered when a token is present in configuration.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task GatewayClientRegisteredWithToken()
    {
        await Assert.That(IsGatewayRegistered("configured-token"))
            .IsTrue();
    }

    /// <summary>
    /// Verifies that no gateway client is registered when the token is absent from configuration.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task GatewayClientNotRegisteredWithoutToken()
    {
        await Assert.That(IsGatewayRegistered(null))
            .IsFalse();
    }

    /// <summary>
    /// Verifies that a whitespace token is treated as absent.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task GatewayClientNotRegisteredForWhitespaceToken()
    {
        await Assert.That(IsGatewayRegistered("   "))
            .IsFalse();
    }

    private static bool IsGatewayRegistered(string? token)
    {
        var builder = Host.CreateApplicationBuilder();

        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            [DiscordExtensions.TokenKey] = token
        });

        builder.ConfigureDiscord();

        return builder.Services.Any(descriptor => descriptor.ServiceType == typeof(GatewayClient));
    }
}
