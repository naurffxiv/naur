// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Naur.Authingway.Data;
using Naur.Authingway.Tests.Testing;

namespace Naur.Authingway.Tests.Data;

/// <summary>
/// Contains tests that verify the application's database context can connect to the underlying database.
/// </summary>
/// <param name="dbContext">The application's database context to be tested. Cannot be null.</param>
[AuthingwayDataSource]
public class DbContextTests(AppDbContext dbContext)
{
    /// <summary>
    /// Verifies that the application's database context can successfully connect to the underlying database.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task CanConnectToDatabase()
    {
        var result = await dbContext.Database.CanConnectAsync();

        await Assert.That(result)
            .IsTrue();
    }
}
