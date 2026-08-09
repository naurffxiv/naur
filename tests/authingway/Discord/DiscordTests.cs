// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Tests.Testing;
using NetCord.Gateway;

namespace Naur.Authingway.Tests.Discord;

/// <summary>
/// Contains tests that verify Discord components are absent from the running application when no token is configured.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
public class DiscordTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that no gateway client is registered when the running application has no configured token.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task GatewayClientNotRegisteredWithoutToken()
    {
        var gatewayClient = serviceProvider.GetService<GatewayClient>();

        await Assert.That(gatewayClient)
            .IsNull();
    }
}
