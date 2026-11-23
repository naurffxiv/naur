// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Configuration.Identity;

namespace Naur.Authingway.Configuration;

/// <summary>
/// Provides extension methods for binding the application's typed configuration.
/// </summary>
public static class ConfigurationExtensions
{
    /// <summary>
    /// Binds the application's typed configuration classes from their configuration sections.
    /// </summary>
    /// <param name="builder">The host application builder to configure. Cannot be null.</param>
    /// <returns>The same <see cref="IHostApplicationBuilder"/> instance for chaining further configuration.</returns>
    /// <remarks>
    /// Invalid configuration fails application startup through strict binding and data-annotation validation.
    /// </remarks>
    public static IHostApplicationBuilder ConfigureConfiguration(this IHostApplicationBuilder builder)
    {
        builder.Services.AddOptions<DiscordConfig>()
            .BindConfiguration("Discord", options => options.ErrorOnUnknownConfiguration = true)
            .ValidateDataAnnotations()
            .ValidateOnStart();

        builder.Services.AddOptions<IdentityConfig>()
            .BindConfiguration("Identity")
            .ValidateDataAnnotations()
            .ValidateOnStart();

        return builder;
    }
}
