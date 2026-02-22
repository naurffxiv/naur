// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Provides integration tests for verifying the health and availability of the OpenAPI documentation endpoint.
/// </summary>
/// <param name="httpClient">The HTTP client used to send requests to the API under test. Must be configured to target the appropriate server instance.</param>
[AuthingwayDataSource]
public class OpenApiTests(HttpClient httpClient)
{
    /// <summary>
    /// Verifies that the OpenAPI document endpoint responds successfully.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task OpenApiDocumentWorking()
    {
        var result = await httpClient.GetAsync("/docs/openapi.json");

        await Assert.That(result.IsSuccessStatusCode)
            .IsTrue();
    }
}
