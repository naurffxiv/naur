// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.Options;

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
    /// Invalid configuration fails application startup with an error naming each offending key.
    /// </remarks>
    public static IHostApplicationBuilder ConfigureConfiguration(this IHostApplicationBuilder builder)
    {
        builder.Services.AddOptions<DiscordConfig>()
            .Bind(builder.Configuration.GetSection("Discord"))
            .ValidateOnStart();
        builder.Services.AddSingleton<IValidateOptions<DiscordConfig>, DiscordConfigValidator>();

        return builder;
    }
}
