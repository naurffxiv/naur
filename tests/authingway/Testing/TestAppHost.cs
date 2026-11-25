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

        await ValidateParametersAsync(Builder.Resources.OfType<ParameterResource>());

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
        if (Application is not null)
        {
            await Application.DisposeAsync();
        }

        if (Builder is not null)
        {
            await Builder.DisposeAsync();
        }

        GC.SuppressFinalize(this);
    }

    private static async Task ValidateParametersAsync(IEnumerable<ParameterResource> parameters)
    {
        List<string> errors = [];

        foreach (var parameter in parameters)
        {
            string reason;
            try
            {
                var value = await parameter.GetValueAsync(CancellationToken.None);
                if (!string.IsNullOrWhiteSpace(value))
                {
                    continue;
                }

                reason = "empty or whitespace";
            }
            catch (MissingParameterValueException)
            {
                reason = "missing (no configured value or default)";
            }

            var configurationKey = parameter.IsConnectionString
                ? $"ConnectionStrings:{parameter.Name}"
                : $"Parameters:{parameter.Name}";

            if (parameter.TryGetLastAnnotation<ResourceSnapshotAnnotation>(out var snapshot))
            {
                configurationKey = snapshot.InitialSnapshot.Properties
                    .FirstOrDefault(property => property.Name == CustomResourceKnownProperties.Source)?.Value as string
                    ?? configurationKey;
            }

            errors.Add($"  - {parameter.Name}: {reason}. Configure '{configurationKey}' "
                + $"(environment variable '{configurationKey.Replace(":", "__")}').");
        }

        if (errors.Count > 0)
        {
            throw new DistributedApplicationException(
                $"AppHost parameter validation failed before starting resources:{Environment.NewLine}"
                + string.Join(Environment.NewLine, errors)
                + $"{Environment.NewLine}Set the environment variables in CI, or configure AppHost user secrets locally:"
                + $"{Environment.NewLine}  dotnet user-secrets set \"<configuration-key>\" \"<value>\" --id naur-apphost"
                + $"{Environment.NewLine}Test runs do not prompt for parameter values.");
        }
    }
}
