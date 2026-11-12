// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

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
            await RunDatabaseMigrationsAsync(stoppingToken);
        }
        catch (Exception ex)
        {
            activity?.AddException(ex);
            throw;
        }
    }

    private async Task RunDatabaseMigrationsAsync(CancellationToken cancellationToken)
    {
        logger.LogDatabaseMigrationStarting();

        while (true)
        {
            try
            {
                using var scope = serviceProvider.CreateScope();

                var migrator = scope.ServiceProvider.GetRequiredService<DatabaseMigrator>();

                if (await migrator.MigrateAsync(cancellationToken) == MigrationOutcome.Applied)
                {
                    break;
                }

                logger.LogDatabaseMigrationAlreadyRunning();
            }
            catch (Exception ex)
            {
                logger.LogDatabaseMigrationFailed(ex);
            }

            await Task.Delay(15000, cancellationToken);
        }

        logger.LogDatabaseMigrationComplete();
    }
}
