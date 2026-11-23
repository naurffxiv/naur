// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore.Design;

namespace Naur.Authingway.Data.Design;

/// <summary>
/// Creates the database context for EF tools and migration bundles.
/// </summary>
public class DesignTimeAppDbContextFactory : IDesignTimeDbContextFactory<AppDbContext>
{
    /// <inheritdoc/>
    public AppDbContext CreateDbContext(string[] args)
    {
        var builder = WebApplication.CreateBuilder(args);

        builder.ConfigureData();

        var app = builder.Build();

        var scope = app.Services.CreateScope();

        return scope.ServiceProvider.GetRequiredService<AppDbContext>();
    }
}
