// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Provides tests to verify the registration and configuration of HttpClient within the service provider.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve dependencies for the tests. Must not be null.</param>
[AuthingwayDataSource]
public class HttpClientTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that an instance of HttpClient is registered in the service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task HttpClientRegistered()
    {
        var httpClient = serviceProvider.GetService<HttpClient>();

        await Assert.That(httpClient)
            .IsNotNull();
    }
}
