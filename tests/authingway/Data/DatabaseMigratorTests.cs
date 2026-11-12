// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Data;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Data;

/// <summary>
/// Contains tests that verify migrations are serialized by the advisory lock.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
[NotInParallel]
public class DatabaseMigratorTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that migrations are applied once the startup migration workers release the lock.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AppliesMigrationsWhenLockIsFree()
    {
        using var scope = serviceProvider.CreateScope();

        var migrator = scope.ServiceProvider.GetRequiredService<DatabaseMigrator>();

        var deadline = DateTime.UtcNow.AddSeconds(30);
        var outcome = await migrator.MigrateAsync(CancellationToken.None);

        while (outcome == MigrationOutcome.AlreadyRunning && DateTime.UtcNow < deadline)
        {
            await Task.Delay(TimeSpan.FromSeconds(1));
            outcome = await migrator.MigrateAsync(CancellationToken.None);
        }

        await Assert.That(outcome)
            .IsEqualTo(MigrationOutcome.Applied);
    }

    /// <summary>
    /// Verifies that a caller reports the migration as already running when the lock is held elsewhere.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReportsAlreadyRunningWhenLockIsHeld()
    {
        var cancellationToken = CancellationToken.None;

        using var holderScope = serviceProvider.CreateScope();

        var holderContext = holderScope.ServiceProvider.GetRequiredService<AppDbContext>();
        var holderConnection = holderContext.Database.GetDbConnection();

        await holderConnection.OpenAsync(cancellationToken);
        await ExecuteLockFunctionAsync(holderContext, "pg_advisory_lock", cancellationToken);

        try
        {
            using var callerScope = serviceProvider.CreateScope();

            var migrator = callerScope.ServiceProvider.GetRequiredService<DatabaseMigrator>();

            var outcome = await migrator.MigrateAsync(cancellationToken);

            await Assert.That(outcome)
                .IsEqualTo(MigrationOutcome.AlreadyRunning);
        }
        finally
        {
            await ExecuteLockFunctionAsync(holderContext, "pg_advisory_unlock", cancellationToken);
        }
    }

    private static async Task ExecuteLockFunctionAsync(AppDbContext dbContext, string function, CancellationToken cancellationToken)
    {
        await using var command = dbContext.Database.GetDbConnection().CreateCommand();

        command.CommandText = $"SELECT {function}({DatabaseMigrator.AdvisoryLockKey})";

        await command.ExecuteScalarAsync(cancellationToken);
    }
}
