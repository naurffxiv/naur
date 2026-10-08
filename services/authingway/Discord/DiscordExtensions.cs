// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using NetCord.Hosting.Gateway;

namespace Naur.Authingway.Discord;

/// <summary>
/// Provides extension methods for configuring Discord integration in an application.
/// </summary>
public static class DiscordExtensions
{
    /// <summary>
    /// Configures the Discord gateway client for the application.
    /// </summary>
    /// <param name="builder">The host application builder to configure. Cannot be null.</param>
    /// <returns>The same <see cref="IHostApplicationBuilder"/> instance for chaining further configuration.</returns>
    public static IHostApplicationBuilder ConfigureDiscord(this IHostApplicationBuilder builder)
    {
        builder.Services.AddDiscordGateway();

        return builder;
    }
}
