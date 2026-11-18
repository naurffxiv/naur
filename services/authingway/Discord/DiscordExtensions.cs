// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using NetCord.Hosting.Gateway;
using NetCord.Hosting.Services;
using NetCord.Hosting.Services.ApplicationCommands;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Discord;

/// <summary>
/// Provides extension methods for configuring Discord integration in an application.
/// </summary>
public static class DiscordExtensions
{
    /// <summary>
    /// Configures the Discord gateway client and application commands when a token is configured.
    /// </summary>
    /// <param name="builder">The host application builder to configure. Cannot be null.</param>
    /// <returns>The same <see cref="IHostApplicationBuilder"/> instance for chaining further configuration.</returns>
    public static IHostApplicationBuilder ConfigureDiscord(this IHostApplicationBuilder builder)
    {
        if (string.IsNullOrWhiteSpace(builder.Configuration["Discord:Token"]))
        {
            return builder;
        }

        builder.Services.AddDiscordGateway(options =>
            {
                options.Intents = default;
            })
            .AddApplicationCommands(options =>
            {
                options.ResultHandler = ApplicationCommandResultHandler<ApplicationCommandContext>.Ephemeral;
            });

        return builder;
    }

    /// <summary>
    /// Registers the Discord command modules when a token is configured.
    /// </summary>
    /// <param name="app">The application to configure. Cannot be null.</param>
    /// <returns>The same <see cref="WebApplication"/> instance for chaining further configuration.</returns>
    public static WebApplication UseDiscord(this WebApplication app)
    {
        if (string.IsNullOrWhiteSpace(app.Configuration["Discord:Token"]))
        {
            return app;
        }

        app.AddModules(typeof(DiscordExtensions).Assembly);

        return app;
    }
}
