// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Provides tests for verifying the behavior of scalar HTTP endpoints using the specified HTTP client.
/// </summary>
/// <param name="httpClient">The HTTP client instance used to send requests to the API endpoints under test. Must not be null.</param>
[AuthingwayDataSource]
public class ScalarTests(HttpClient httpClient)
{
    /// <summary>
    /// Verifies that a GET request to the "/docs" endpoint returns a successful HTTP status code.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ScalarWorking()
    {
        var result = await httpClient.GetAsync("/docs");

        await Assert.That(result.IsSuccessStatusCode)
            .IsTrue();
    }
}
