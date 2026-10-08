// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.AspNetCore.Identity.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore;
using Naur.Authingway.Data.Entities;

namespace Naur.Authingway.Data;

/// <summary>
/// Represents the Entity Framework Core database context for the application.
/// </summary>
/// <param name="options">The options to be used by the DbContext. Must not be null.</param>
public class AppDbContext(DbContextOptions<AppDbContext> options) : IdentityDbContext<User, Role, Guid, UserClaim, UserRole, UserLogin, RoleClaim, UserToken, UserPasskey>(options)
{
    /// <inheritdoc/>
    protected override void OnModelCreating(ModelBuilder builder)
    {
        base.OnModelCreating(builder);

        builder.ApplyConfigurationsFromAssembly(typeof(AppDbContext).Assembly);
    }
}
