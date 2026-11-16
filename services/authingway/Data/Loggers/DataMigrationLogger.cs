// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Data.Loggers;

/// <summary>
/// Provides strongly-typed log messages for database migration runs.
/// </summary>
public static partial class DataMigrationLogger
{
    /// <summary>
    /// Logs that a migration run is starting.
    /// </summary>
    /// <param name="logger">The logger to write to.</param>
    /// <param name="executor">Who started the migration.</param>
    [LoggerMessage(1001, LogLevel.Information, "Starting database migration for {Executor}...")]
    public static partial void LogDatabaseMigrationStarting(this ILogger logger, string executor);

    /// <summary>
    /// Logs that a migration run failed.
    /// </summary>
    /// <param name="logger">The logger to write to.</param>
    /// <param name="executor">Who started the migration.</param>
    /// <param name="exception">The error that ended the run.</param>
    [LoggerMessage(1002, LogLevel.Error, "Database migration for {Executor} failed.")]
    public static partial void LogDatabaseMigrationFailed(this ILogger logger, string executor, Exception exception);

    /// <summary>
    /// Logs that a migration run applied its pending migrations.
    /// </summary>
    /// <param name="logger">The logger to write to.</param>
    /// <param name="executor">Who started the migration.</param>
    /// <param name="count">The number of migrations applied.</param>
    [LoggerMessage(1003, LogLevel.Information, "Database migration for {Executor} complete, {Count} migrations applied.")]
    public static partial void LogDatabaseMigrationComplete(this ILogger logger, string executor, int count);

    /// <summary>
    /// Logs that a migration run was refused because another run is in progress.
    /// </summary>
    /// <param name="logger">The logger to write to.</param>
    /// <param name="executor">Who attempted the migration.</param>
    [LoggerMessage(1004, LogLevel.Information, "Database migration for {Executor} skipped, another migration is running.")]
    public static partial void LogDatabaseMigrationAlreadyRunning(this ILogger logger, string executor);

    /// <summary>
    /// Logs that a migration run found no pending migrations.
    /// </summary>
    /// <param name="logger">The logger to write to.</param>
    /// <param name="executor">Who started the migration.</param>
    [LoggerMessage(1005, LogLevel.Information, "Database migration for {Executor} found the database up to date.")]
    public static partial void LogDatabaseMigrationUpToDate(this ILogger logger, string executor);
}
