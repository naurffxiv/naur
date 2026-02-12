// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Diagnostics;
using Naur.Authingway.Data.Extensions;
using Naur.Authingway.Data.Workers;

namespace Naur.Authingway.Data;

/// <summary>
/// Provides extension methods for configuring data-related services in an application.
/// </summary>
public static class DataExtensions
{
    /// <summary>
    /// Configures data-related services for the application using the specified host application builder.
    /// </summary>
    /// <param name="builder">The host application builder to configure with data services. Cannot be null.</param>
    /// <returns>The same instance of <see cref="IHostApplicationBuilder"/> with data services configured.</returns>
    public static IHostApplicationBuilder ConfigureData(this IHostApplicationBuilder builder)
    {
        return builder.ConfigureDbContext()
            .ConfigureMigrationWorker();
    }

    private static IHostApplicationBuilder ConfigureDbContext(this IHostApplicationBuilder builder)
    {
        builder.AddNpgsqlDbContext<AppDbContext>("authingwaydb", configureDbContextOptions: options =>
        {
            if (builder.Environment.IsDevelopment())
            {
                options.EnableSensitiveDataLogging();
            }

            options.ConfigureWarnings(config =>
            {
                config.TreatWarningsAsErrors();

                if (builder.Environment.IsDevelopment())
                {
                    config.Ignore(CoreEventId.SensitiveDataLoggingEnabledWarning);
                }
            });
        });

        return builder;
    }

    private static IHostApplicationBuilder ConfigureMigrationWorker(this IHostApplicationBuilder builder)
    {
        builder.Services.AddHostedService<DataMigrationWorker>();

        return builder;
    }
}
