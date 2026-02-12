// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using FastEndpoints;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.Extensions.Diagnostics.HealthChecks;
using OpenTelemetry;
using OpenTelemetry.Metrics;
using OpenTelemetry.Trace;
using Quartz;
using Scalar.AspNetCore;
using System.Reflection;

namespace Naur.Authingway.Core;

/// <summary>
/// Provides extension methods for configuring core services and middleware in an ASP.NET Core application, including
/// telemetry, health checks, service discovery, HTTP client configuration, OpenAPI documentation, and FastEndpoints
/// integration.
/// </summary>
public static class CoreExtensions
{
    /// <summary>
    /// Configures core services and middleware for the application, including telemetry, health checks, service
    /// discovery, HTTP client, OpenAPI, and FastEndpoints.
    /// </summary>
    /// <param name="builder">The application builder to configure. Cannot be null.</param>
    /// <returns>The same IHostApplicationBuilder instance for chaining further configuration.</returns>
    public static IHostApplicationBuilder ConfigureCore(this IHostApplicationBuilder builder)
    {
        return builder.ConfigureTelemetry()
            .ConfigureHealthChecks()
            .ConfigureServiceDiscovery()
            .ConfigureHttpClient()
            .ConfigureQuartz()
            .ConfigureSecurity()
            .ConfigureOpenApi()
            .ConfigureFastEndpoints();
    }

    /// <summary>
    /// Configures the specified web application with core middleware components, including health checks, error pages,
    /// security features, OpenAPI documentation, Scalar endpoints, and FastEndpoints support.
    /// </summary>
    /// <param name="app">The <see cref="WebApplication"/> instance to configure with core middleware components. Cannot be null.</param>
    /// <returns>The <see cref="WebApplication"/> instance with the core middleware components configured.</returns>
    public static WebApplication UseCore(this WebApplication app)
    {
        return app.UseHealthChecks()
            .UseErrorPages()
            .UseSecurity()
            .UseOpenApi()
            .UseScalar()
            .UseFastEndpoints();
    }

    private static IHostApplicationBuilder ConfigureTelemetry(this IHostApplicationBuilder builder)
    {
        builder.Logging.AddOpenTelemetry(logging =>
        {
            logging.IncludeFormattedMessage = true;
            logging.IncludeScopes = true;
        });

        builder.Services.AddOpenTelemetry()
            .WithMetrics(metrics =>
            {
                metrics.AddRuntimeInstrumentation()
                    .AddAspNetCoreInstrumentation()
                    .AddHttpClientInstrumentation();
            })
            .WithTracing(tracing =>
            {
                tracing.AddSource(builder.Environment.ApplicationName)
                    .AddAspNetCoreInstrumentation(tracing =>
                        tracing.Filter = context =>
                            !context.Request.Path.StartsWithSegments("/health")
                            && !context.Request.Path.StartsWithSegments("/alive")
                    )
                    .AddHttpClientInstrumentation()
                    .AddEntityFrameworkCoreInstrumentation()
                    .AddSource("Naur.Authingway.Data");
            });

        if (!string.IsNullOrWhiteSpace(builder.Configuration["OTEL_EXPORTER_OTLP_ENDPOINT"]))
        {
            builder.Services.AddOpenTelemetry()
                .UseOtlpExporter();
        }

        return builder;
    }

    private static IHostApplicationBuilder ConfigureHealthChecks(this IHostApplicationBuilder builder)
    {
        builder.Services.AddHealthChecks()
            .AddCheck("self", () => HealthCheckResult.Healthy(), ["live"]);

        return builder;
    }

    private static IHostApplicationBuilder ConfigureServiceDiscovery(this IHostApplicationBuilder builder)
    {
        builder.Services.AddServiceDiscovery();

        return builder;
    }

    private static IHostApplicationBuilder ConfigureHttpClient(this IHostApplicationBuilder builder)
    {
        builder.Services.ConfigureHttpClientDefaults(http =>
        {
            http.AddStandardResilienceHandler();
            http.AddServiceDiscovery();
        });

        return builder;
    }

    private static IHostApplicationBuilder ConfigureQuartz(this IHostApplicationBuilder builder)
    {
        builder.Services.AddQuartz(options =>
        {
            options.UseSimpleTypeLoader();
            options.UseInMemoryStore();
        });

        builder.Services.AddQuartzHostedService(options => options.WaitForJobsToComplete = true);

        return builder;
    }

    private static IHostApplicationBuilder ConfigureSecurity(this IHostApplicationBuilder builder)
    {
        builder.Services.AddAuthentication();
        builder.Services.AddAuthorization();

        return builder;
    }

    private static IHostApplicationBuilder ConfigureOpenApi(this IHostApplicationBuilder builder)
    {
        builder.Services.AddOpenApi(options =>
        {
            options.AddDocumentTransformer((document, context, cancellationToken) =>
            {
                document.Info.Title = "Authingway";
                document.Info.Description = "The authentication and authorization server for NAUR.";
                document.Info.Version = Assembly.GetEntryAssembly()!.GetName().Version!.ToString(3);

                document.Info.Contact = new()
                {
                    Name = "NAUR",
                    Email = "naurffxiv@gmail.com"
                };

                return Task.CompletedTask;
            });
        });

        return builder;
    }

    private static IHostApplicationBuilder ConfigureFastEndpoints(this IHostApplicationBuilder builder)
    {
        builder.Services.AddFastEndpoints();

        return builder;
    }

    private static WebApplication UseHealthChecks(this WebApplication app)
    {
        app.MapHealthChecks("/health");

        app.MapHealthChecks("/alive", new HealthCheckOptions
        {
            Predicate = r => r.Tags.Contains("live")
        });

        return app;
    }

    private static WebApplication UseErrorPages(this WebApplication app)
    {
        if (app.Environment.IsDevelopment())
        {
            app.UseDeveloperExceptionPage();
        }

        return app;
    }

    private static WebApplication UseSecurity(this WebApplication app)
    {
        if (!app.Environment.IsDevelopment())
        {
            app.UseHsts();
        }

        app.UseHttpsRedirection();

        app.UseAuthentication();
        app.UseAuthorization();

        return app;
    }

    private static WebApplication UseOpenApi(this WebApplication app)
    {
        app.MapOpenApi("/docs/openapi.json");

        return app;
    }

    private static WebApplication UseScalar(this WebApplication app)
    {
        app.MapScalarApiReference("/docs", options =>
        {
            options.WithTitle("Authingway");
            options.WithOpenApiRoutePattern("/docs/openapi.json");
            options.HideClientButton = true;
        });

        return app;
    }

    private static WebApplication UseFastEndpoints(this WebApplication app)
    {
        app.MapFastEndpoints();

        return app;
    }
}
