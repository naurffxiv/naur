// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.Authorization;
using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Provides integration tests for verifying the health and availability of security services.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
public class SecurityTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that the authentication service is registered in the service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AuthenticationServiceRegistered()
    {
        var authenticationService = serviceProvider.GetService<IAuthenticationService>();

        await Assert.That(authenticationService)
            .IsNotNull();
    }

    /// <summary>
    /// Verifies that the authorization service is registered in the service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AuthorizationServiceRegistered()
    {
        var authorizationService = serviceProvider.GetService<IAuthorizationService>();

        await Assert.That(authorizationService)
            .IsNotNull();
    }
}
