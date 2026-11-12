// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using NetCord.Hosting.Gateway;
using NetCord.Hosting.Services;
using NetCord.Hosting.Services.ApplicationCommands;

namespace Naur.Authingway.Discord;

/// <summary>
/// Provides extension methods for configuring Discord integration in an application.
/// </summary>
public static class DiscordExtensions
{
    /// <summary>
    /// Configures the Discord gateway client and application commands for the application.
    /// </summary>
    /// <param name="builder">The host application builder to configure. Cannot be null.</param>
    /// <returns>The same <see cref="IHostApplicationBuilder"/> instance for chaining further configuration.</returns>
    public static IHostApplicationBuilder ConfigureDiscord(this IHostApplicationBuilder builder)
    {
        builder.Services.Configure<DiscordOptions>(builder.Configuration.GetSection(DiscordOptions.SectionName));
        builder.Services.AddSingleton<MigrationCommandAuthorizer>();

        builder.Services.AddDiscordGateway()
            .AddApplicationCommands();

        return builder;
    }

    /// <summary>
    /// Registers the Discord command modules for the application.
    /// </summary>
    /// <param name="app">The application to configure. Cannot be null.</param>
    /// <returns>The same <see cref="WebApplication"/> instance for chaining further configuration.</returns>
    public static WebApplication UseDiscord(this WebApplication app)
    {
        app.AddModules(typeof(DiscordExtensions).Assembly);

        return app;
    }
}
