// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Data.Loggers;

/// <summary>
/// Provides extension methods for logging database migration events using strongly-typed log messages.
/// </summary>
public static partial class DataMigrationLogger
{
    [LoggerMessage(1001, LogLevel.Information, "Starting database migration...")]
    public static partial void LogDatabaseMigrationStarting(this ILogger logger);

    [LoggerMessage(1002, LogLevel.Warning, "Failed to migrate database, retrying in 15 seconds...")]
    public static partial void LogDatabaseMigrationFailed(this ILogger logger, Exception ex);

    [LoggerMessage(1003, LogLevel.Information, "Database migration complete!")]
    public static partial void LogDatabaseMigrationComplete(this ILogger logger);
}
