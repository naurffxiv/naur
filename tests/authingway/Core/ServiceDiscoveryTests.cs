// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.ServiceDiscovery;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Contains tests that verify service discovery components are correctly registered in the dependency injection
/// container.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
public class ServiceDiscoveryTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that the ServiceEndpointResolver service is registered in the dependency injection container.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ServiceEndpointResolverRegistered()
    {
        var endpointResolver = serviceProvider.GetService<ServiceEndpointResolver>();

        await Assert.That(endpointResolver)
            .IsNotNull();
    }
}
