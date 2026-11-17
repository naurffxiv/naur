// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Naur.Authingway.Data.Loggers;
using System.Diagnostics;

namespace Naur.Authingway.Data;

/// <summary>
/// Describes the result of an attempt to apply database migrations.
/// </summary>
public enum MigrationOutcome
{
    /// <summary>
    /// Pending migrations were applied.
    /// </summary>
    Applied,

    /// <summary>
    /// There were no pending migrations to apply.
    /// </summary>
    UpToDate,

    /// <summary>
    /// Another migration was running in this process, so no migrations were attempted.
    /// </summary>
    AlreadyRunning,

    /// <summary>
    /// The migration attempt raised an error, which was logged.
    /// </summary>
    Failed
}

/// <summary>
/// Applies pending database migrations, one run at a time.
/// </summary>
/// <param name="dbContext">The database context whose migrations are applied. Cannot be null.</param>
/// <param name="gate">Admits one migration at a time within this process. Cannot be null.</param>
/// <param name="logger">Records each migration run. Cannot be null.</param>
public sealed class DatabaseMigrator(AppDbContext dbContext, MigrationGate gate, ILogger<DatabaseMigrator> logger)
{
    private static readonly ActivitySource s_activitySource = new("Naur.Authingway.Data");

    /// <summary>
    /// Applies any pending migrations, unless another migration is already running in this process.
    /// </summary>
    /// <param name="executor">Who started the migration, recorded in logs and traces. Cannot be null.</param>
    /// <param name="cancellationToken">A token that cancels the operation.</param>
    /// <returns>The outcome of the attempt.</returns>
    /// <exception cref="OperationCanceledException"><paramref name="cancellationToken"/> was canceled.</exception>
    /// <remarks>
    /// EF Core takes its own database lock while migrating, which serializes migrations across processes. Errors
    /// other than cancellation are logged and reported as <see cref="MigrationOutcome.Failed"/>.
    /// </remarks>
    public async Task<MigrationOutcome> MigrateAsync(string executor, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(executor);

        if (!gate.TryEnter())
        {
            logger.LogDatabaseMigrationAlreadyRunning(executor);

            return MigrationOutcome.AlreadyRunning;
        }

        try
        {
            using var activity = s_activitySource.StartActivity("Migrating data", ActivityKind.Client);
            activity?.SetTag("migration.executor", executor);

            logger.LogDatabaseMigrationStarting(executor);

            try
            {
                var pending = (await dbContext.Database.GetPendingMigrationsAsync(cancellationToken)).Count();

                if (pending == 0)
                {
                    logger.LogDatabaseMigrationUpToDate(executor);

                    return MigrationOutcome.UpToDate;
                }

                await dbContext.Database.MigrateAsync(cancellationToken);

                logger.LogDatabaseMigrationComplete(executor, pending);

                return MigrationOutcome.Applied;
            }
            catch (Exception ex) when (ex is not OperationCanceledException)
            {
                activity?.AddException(ex);
                logger.LogDatabaseMigrationFailed(executor, ex);

                return MigrationOutcome.Failed;
            }
        }
        finally
        {
            gate.Exit();
        }
    }
}
