// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging.Abstractions;
using Naur.Authingway.Data;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Data;

/// <summary>
/// Contains tests that verify database migrations run one at a time and report their outcome.
/// </summary>
/// <param name="serviceProvider">The service provider used to resolve registered services for testing.</param>
[AuthingwayDataSource]
[NotInParallel]
public class DatabaseMigratorTests(IServiceProvider serviceProvider)
{
    /// <summary>
    /// Verifies that a migration reports the database as up to date once the test host has migrated it.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReportsUpToDateWhenNothingIsPending()
    {
        using var scope = serviceProvider.CreateScope();

        var migrator = scope.ServiceProvider.GetRequiredService<DatabaseMigrator>();

        var outcome = await migrator.MigrateAsync("test", CancellationToken.None);

        await Assert.That(outcome)
            .IsEqualTo(MigrationOutcome.UpToDate);
    }

    /// <summary>
    /// Verifies that a migration reports another migration as running while the gate is held.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReportsAlreadyRunningWhileTheGateIsHeld()
    {
        var gate = serviceProvider.GetRequiredService<MigrationGate>();

        await Assert.That(gate.TryEnter()).IsTrue();

        try
        {
            using var scope = serviceProvider.CreateScope();

            var migrator = scope.ServiceProvider.GetRequiredService<DatabaseMigrator>();

            var outcome = await migrator.MigrateAsync("test", CancellationToken.None);

            await Assert.That(outcome)
                .IsEqualTo(MigrationOutcome.AlreadyRunning);
        }
        finally
        {
            gate.Exit();
        }
    }

    /// <summary>
    /// Verifies that a migration that cannot reach the database reports failure and releases the gate.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReportsFailedWhenDatabaseIsUnreachable()
    {
        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseNpgsql("Host=127.0.0.1;Port=1;Database=authingway;Username=postgres;Password=unused;Timeout=2")
            .Options;

        await using var dbContext = new AppDbContext(options);
        using var gate = new MigrationGate();

        var migrator = new DatabaseMigrator(dbContext, gate, NullLogger<DatabaseMigrator>.Instance);

        var outcome = await migrator.MigrateAsync("test", CancellationToken.None);

        await Assert.That(outcome).IsEqualTo(MigrationOutcome.Failed);
        await Assert.That(gate.TryEnter()).IsTrue();
    }

    /// <summary>
    /// Verifies that a canceled migration throws and still releases the gate.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReleasesTheGateWhenCanceled()
    {
        using var scope = serviceProvider.CreateScope();

        var migrator = scope.ServiceProvider.GetRequiredService<DatabaseMigrator>();
        var gate = serviceProvider.GetRequiredService<MigrationGate>();

        using var canceled = new CancellationTokenSource();
        await canceled.CancelAsync();

        await Assert.That(() => migrator.MigrateAsync("test", canceled.Token))
            .Throws<OperationCanceledException>();

        await Assert.That(gate.TryEnter()).IsTrue();
        gate.Exit();
    }
}
