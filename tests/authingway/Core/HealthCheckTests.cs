// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Contains tests that verify the health and liveness endpoints respond as expected.
/// </summary>
/// <param name="httpClient">The HTTP client used to send requests to the service under test. Must be configured to target the correct service
/// instance.</param>
[AuthingwayDataSource]
public class HealthCheckTests(HttpClient httpClient)
{
    /// <summary>
    /// Verifies that the health endpoint returns a successful response indicating the service is healthy.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task HealthEndpointWorking()
    {
        var result = await httpClient.GetStringAsync("/health");

        await Assert.That(result)
            .IsEqualTo("Healthy");
    }

    /// <summary>
    /// Verifies that the /alive endpoint responds with a healthy status.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AliveEndpointWorking()
    {
        var result = await httpClient.GetStringAsync("/alive");

        await Assert.That(result)
            .IsEqualTo("Healthy");
    }
}
