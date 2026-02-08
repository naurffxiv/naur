// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Aspire.Hosting.Testing;
using Microsoft.Extensions.DependencyInjection;

namespace Naur.Authingway.Tests.Testing;

/// <summary>
/// Provides a data source for test data generation that resolves services from the Authingway dependency injection container.
/// </summary>
public class AuthingwayDataSourceAttribute : DependencyInjectionDataSourceAttribute<IServiceScope>
{
    /// <summary>
    /// Gets the web application factory used to create test server instances for integration testing.
    /// </summary>
    [ClassDataSource<TestWebApplicationFactory>(Shared = SharedType.PerTestSession)]
    public TestWebApplicationFactory Factory { get; init; } = default!;

    /// <inheritdoc/>
    public override IServiceScope CreateScope(DataGeneratorMetadata dataGeneratorMetadata)
    {
        return Factory.Server.Services.CreateAsyncScope();
    }

    /// <inheritdoc/>
    public override object? Create(IServiceScope scope, Type type)
    {
        if (type == typeof(HttpClient))
        {
            return Factory.AppHost.Application.CreateHttpClient("authingway");
        }

        return scope.ServiceProvider.GetService(type);
    }
}
