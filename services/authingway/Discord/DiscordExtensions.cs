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
    /// The configuration key holding the Discord bot token.
    /// </summary>
    public const string TokenKey = "Discord:Token";

    /// <summary>
    /// Configures the Discord gateway client for the application.
    /// </summary>
    /// <param name="builder">The host application builder to configure. Cannot be null.</param>
    /// <returns>The same <see cref="IHostApplicationBuilder"/> instance for chaining further configuration.</returns>
    /// <remarks>
    /// The gateway is registered only when <see cref="TokenKey"/> resolves to a non-whitespace value. When no token is
    /// configured the application runs with no gateway client registered and opens no connection to Discord.
    /// </remarks>
    public static IHostApplicationBuilder ConfigureDiscord(this IHostApplicationBuilder builder)
    {
        if (!builder.HasDiscordToken())
        {
            return builder;
        }

        builder.Services.AddDiscordGateway();

        return builder;
    }

    private static bool HasDiscordToken(this IHostApplicationBuilder builder)
    {
        return !string.IsNullOrWhiteSpace(builder.Configuration[TokenKey]);
    }
}
