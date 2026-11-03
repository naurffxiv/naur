// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.AspNetCore.Identity;
using Naur.Authingway.Data;
using Naur.Authingway.Data.Entities;

namespace Naur.Authingway.Identity;

/// <summary>
/// Provides extension methods for configuring identity services in an application.
/// </summary>
public static class IdentityExtensions
{
    /// <summary>
    /// Configures identity services for the application using the specified host application builder.
    /// </summary>
    /// <param name="builder">The host application builder to configure identity services for. Cannot be null.</param>
    /// <returns>The same instance of <see cref="IHostApplicationBuilder"/> with identity services configured.</returns>
    public static IHostApplicationBuilder ConfigureIdentity(this IHostApplicationBuilder builder)
    {
        return builder.ConfigureAspNetCoreIdentity();
    }

    private static IHostApplicationBuilder ConfigureAspNetCoreIdentity(this IHostApplicationBuilder builder)
    {
        builder.Services.AddIdentity<User, Role>(options => options.Stores.SchemaVersion = IdentitySchemaVersions.Version3)
            .AddEntityFrameworkStores<AppDbContext>()
            .AddDefaultTokenProviders();

        return builder;
    }
}
