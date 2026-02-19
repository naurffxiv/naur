// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Aspire.Hosting;
using Aspire.Hosting.ApplicationModel;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.Extensions.Configuration;
using TUnit.Core.Interfaces;

namespace Naur.Authingway.Tests.Testing;

/// <summary>
/// Provides a custom web application factory for integration testing, enabling configuration of the test application host and environment variables.
/// </summary>
public class TestWebApplicationFactory : WebApplicationFactory<Program>, IAsyncDiscoveryInitializer
{
    /// <summary>
    /// Gets the test application host instance used for integration testing.
    /// </summary>
    [ClassDataSource<TestAppHost>(Shared = SharedType.PerTestSession)]
    public required TestAppHost AppHost { get; init; }

    private readonly Dictionary<string, string?> _environmentVariables = [];

    /// <inheritdoc/>
    public async Task InitializeAsync()
    {
        var authingway = (IResourceWithEnvironment)AppHost.Builder.Resources.First(r => r.Name == "authingway");

        var options = new DistributedApplicationExecutionContextOptions(DistributedApplicationOperation.Run)
        {
            ServiceProvider = AppHost.Application.Services
        };

        var executionContext = new DistributedApplicationExecutionContext(options);

        var executionConfiguration = await ExecutionConfigurationBuilder.Create(authingway)
            .WithEnvironmentVariablesConfig()
            .BuildAsync(executionContext);

        foreach (var variable in executionConfiguration.EnvironmentVariables)
        {
            _environmentVariables[variable.Key.Replace("__", ":")] = variable.Value;
        }

        StartServer();
    }

    /// <inheritdoc/>
    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(_environmentVariables)
            .Build();

        builder.UseConfiguration(configuration);
    }
}
