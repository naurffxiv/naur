// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Naur.Authingway.Tests.Testing;
using OpenTelemetry.Metrics;
using OpenTelemetry.Trace;

namespace Naur.Authingway.Tests.Core;

/// <summary>
/// Contains unit tests for verifying telemetry-related service registrations in the application's dependency injection
/// container.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve telemetry services for testing.</param>
[AuthingwayDataSource]
public class TelemetryTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that an ILogger service is registered and can be resolved from the service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task LoggerRegistered()
    {
        var logger = serviceProvider.GetService<ILogger<TelemetryTests>>();

        await Assert.That(logger)
            .IsNotNull();
    }

    /// <summary>
    /// Verifies that a MeterProvider instance is registered and can be resolved from the service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task MeterProviderRegistered()
    {
        var meter = serviceProvider.GetService<MeterProvider>();

        await Assert.That(meter)
            .IsNotNull();
    }

    /// <summary>
    /// Verifies that a TracerProvider instance is registered in the service provider.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task TracerProviderRegistered()
    {
        var tracer = serviceProvider.GetService<TracerProvider>();

        await Assert.That(tracer)
            .IsNotNull();
    }
}
