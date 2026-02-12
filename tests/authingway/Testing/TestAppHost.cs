// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Aspire.Hosting;
using Aspire.Hosting.ApplicationModel;
using Aspire.Hosting.Testing;
using Microsoft.Extensions.DependencyInjection;
using TUnit.Core.Interfaces;

namespace Naur.Authingway.Tests.Testing;

/// <summary>
/// Provides a host for initializing and managing distributed application testing services and their lifecycle within a test context.
/// </summary>
public class TestAppHost : IAsyncDiscoveryInitializer, IAsyncDisposable
{
    /// <summary>
    /// Gets the builder used to configure distributed application testing services.
    /// </summary>
    public IDistributedApplicationTestingBuilder Builder { get; private set; } = default!;

    /// <summary>
    /// Gets the distributed application associated with the current context.
    /// </summary>
    public DistributedApplication Application { get; private set; } = default!;

    /// <inheritdoc/>
    public async Task InitializeAsync()
    {
        Builder = await DistributedApplicationTestingBuilder.CreateAsync<Projects.Naur_AppHost>();

        Builder.Services.ConfigureHttpClientDefaults(http =>
        {
            http.AddStandardResilienceHandler();
        });

        Application = await Builder.BuildAsync();

        var resourceNotificationService = Application.Services.GetRequiredService<ResourceNotificationService>();

        await Application.StartAsync();

        await resourceNotificationService.WaitForResourceHealthyAsync("authingway")
            .WaitAsync(TimeSpan.FromMinutes(5));
    }

    /// <inheritdoc/>
    public async ValueTask DisposeAsync()
    {
        await Application.DisposeAsync();
        await Builder.DisposeAsync();

        GC.SuppressFinalize(this);
    }
}
