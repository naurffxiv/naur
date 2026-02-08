// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Naur.Authingway.Data.Loggers;
using System.Diagnostics;

namespace Naur.Authingway.Data.Workers;

/// <summary>
/// A background service that performs database migrations at application startup.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve required services for performing the migration.</param>
/// <param name="logger">The logger used to record migration progress and errors.</param>
public class DataMigrationWorker(IServiceProvider serviceProvider, ILogger<DataMigrationWorker> logger) : BackgroundService
{
    private static readonly ActivitySource s_activitySource = new("Naur.Authingway.Data");

    /// <inheritdoc/>
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        using var activity = s_activitySource.StartActivity("Migrating data", ActivityKind.Client);

        try
        {
            using var scope = serviceProvider.CreateScope();

            var dbContext = scope.ServiceProvider.GetRequiredService<AppDbContext>();
            await RunDatabaseMigrationsAsync(dbContext, stoppingToken);
        }
        catch (Exception ex)
        {
            activity?.AddException(ex);
            throw;
        }
    }

    private async Task RunDatabaseMigrationsAsync(AppDbContext dbContext, CancellationToken cancellationToken)
    {
        logger.LogDatabaseMigrationStarting();

        while (true)
        {
            try
            {
                var strategy = dbContext.Database.CreateExecutionStrategy();

                await strategy.ExecuteAsync(async () =>
                {
                    await dbContext.Database.MigrateAsync(cancellationToken);
                });
            }
            catch (Exception ex)
            {
                logger.LogDatabaseMigrationFailed(ex);
                await Task.Delay(15000, cancellationToken);
                continue;
            }

            break;
        }

        logger.LogDatabaseMigrationComplete();
    }
}
