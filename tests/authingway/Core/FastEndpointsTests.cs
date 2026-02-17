// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using FastEndpoints;
using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Provides integration tests for FastEndpoints to verify that required services, such as IEndpointFactory, are
/// registered and can be resolved from the dependency injection container.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve dependencies required for testing FastEndpoints functionality.</param>
[AuthingwayDataSource]
public class FastEndpointsTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that an implementation of the IEndpointFactory interface is registered and can be resolved from the
    /// service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task EndpointFactoryRegistered()
    {
        var endpointFactory = serviceProvider.GetService<IEndpointFactory>();

        await Assert.That(endpointFactory)
            .IsNotNull();
    }
}
