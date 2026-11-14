// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using System.Data;
using System.Data.Common;

namespace Naur.Authingway.Data;

/// <summary>
/// Describes the result of an attempt to apply database migrations.
/// </summary>
public enum MigrationOutcome
{
    /// <summary>
    /// Migrations were applied.
    /// </summary>
    Applied,

    /// <summary>
    /// Another caller held the migration lock, so no migrations were attempted.
    /// </summary>
    AlreadyRunning,

    /// <summary>
    /// The migration attempt raised an error, which was logged.
    /// </summary>
    Failed
}

/// <summary>
/// Applies database migrations under an advisory lock so that only one caller migrates at a time.
/// </summary>
/// <param name="dbContext">The database context whose migrations are applied. Cannot be null.</param>
/// <param name="logger">Records failed migration attempts. Cannot be null.</param>
public sealed partial class DatabaseMigrator(AppDbContext dbContext, ILogger<DatabaseMigrator> logger)
{
    /// <summary>
    /// The PostgreSQL advisory lock key that serializes migration runs.
    /// </summary>
    public const long AdvisoryLockKey = 4861710;

    /// <summary>
    /// Applies any pending migrations, unless another caller currently holds the migration lock.
    /// </summary>
    /// <param name="cancellationToken">A token that cancels the operation.</param>
    /// <returns>
    /// <see cref="MigrationOutcome.Applied"/> when migrations ran, <see cref="MigrationOutcome.AlreadyRunning"/>
    /// when the lock was already held, or <see cref="MigrationOutcome.Failed"/> when connecting, locking or migrating
    /// raised an error.
    /// </returns>
    /// <exception cref="OperationCanceledException"><paramref name="cancellationToken"/> was canceled.</exception>
    /// <remarks>
    /// The lock is a PostgreSQL session advisory lock held on this context's connection, so it excludes callers in
    /// other processes as well as other threads. Acquisition does not wait: a caller that cannot take the lock returns
    /// <see cref="MigrationOutcome.AlreadyRunning"/> immediately. Errors other than cancellation are logged rather
    /// than thrown.
    /// </remarks>
    public async Task<MigrationOutcome> MigrateAsync(CancellationToken cancellationToken)
    {
        try
        {
            var connection = dbContext.Database.GetDbConnection();

            if (connection.State != ConnectionState.Open)
            {
                await connection.OpenAsync(cancellationToken);
            }

            if (!await TryAcquireLockAsync(connection, cancellationToken))
            {
                return MigrationOutcome.AlreadyRunning;
            }

            try
            {
                await dbContext.Database.MigrateAsync(cancellationToken);
            }
            finally
            {
                await ReleaseLockAsync(connection, cancellationToken);
            }

            return MigrationOutcome.Applied;
        }
        catch (Exception ex) when (ex is not OperationCanceledException)
        {
            LogMigrationFailed(logger, ex);

            return MigrationOutcome.Failed;
        }
    }

    [LoggerMessage(1, LogLevel.Error, "Database migration failed.")]
    private static partial void LogMigrationFailed(ILogger logger, Exception exception);

    private static async Task<bool> TryAcquireLockAsync(DbConnection connection, CancellationToken cancellationToken)
    {
        var acquired = await ExecuteLockFunctionAsync(connection, "pg_try_advisory_lock", cancellationToken);

        return acquired;
    }

    private static async Task ReleaseLockAsync(DbConnection connection, CancellationToken cancellationToken)
    {
        await ExecuteLockFunctionAsync(connection, "pg_advisory_unlock", cancellationToken);
    }

    private static async Task<bool> ExecuteLockFunctionAsync(DbConnection connection, string function, CancellationToken cancellationToken)
    {
        await using var command = connection.CreateCommand();

        command.CommandText = $"SELECT {function}(@key)";

        var parameter = command.CreateParameter();
        parameter.ParameterName = "key";
        parameter.Value = AdvisoryLockKey;
        command.Parameters.Add(parameter);

        var result = await command.ExecuteScalarAsync(cancellationToken);

        return result is true;
    }
}
