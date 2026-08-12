// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Tests.Testing;
using NetCord.Gateway;

namespace Naur.Authingway.Tests.Discord;

/// <summary>
/// Contains tests that verify Discord components are correctly registered in the dependency injection container.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
public class DiscordTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that the GatewayClient service is registered in the dependency injection container.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task GatewayClientRegistered()
    {
        var gatewayClient = serviceProvider.GetService<GatewayClient>();

        await Assert.That(gatewayClient)
            .IsNotNull();
    }
}
