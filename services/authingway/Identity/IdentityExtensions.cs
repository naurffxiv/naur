// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Data;
using Naur.Authingway.Data.Entities;
using Naur.Authingway.Configuration.Identity;
using Naur.Authingway.Configuration.Discord;
using OpenIddict.Abstractions;

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
        return builder.ConfigureOpenIddict();
    }

    private static IHostApplicationBuilder ConfigureOpenIddict(this IHostApplicationBuilder builder)
    {
        var identityConfig = builder.Configuration.GetSection("Identity")
            .Get<IdentityConfig>() ?? new IdentityConfig();

        var discordConfig = builder.Configuration.GetSection("Discord")
            .Get<DiscordConfig>() ?? new DiscordConfig();

        builder.Services.AddOpenIddict()
            .AddCore(options =>
            {
                options.UseEntityFrameworkCore()
                    .UseDbContext<AppDbContext>()
                    .ReplaceDefaultEntities<Application, Authorization, Scope, Token, Guid>();

                options.UseQuartz();
            })
            .AddClient(options =>
            {
                options.AllowAuthorizationCodeFlow()
                    .AllowRefreshTokenFlow();

                if (identityConfig.IsEmpty && builder.Environment.IsDevelopment())
                {
                    options.AddDevelopmentEncryptionCertificate()
                        .AddDevelopmentSigningCertificate();
                }
                else
                {
                    options.AddSigningCertificate(identityConfig.LoadSigningCertificate())
                        .AddEncryptionCertificate(identityConfig.LoadEncryptionCertificate());
                }

                options.UseAspNetCore()
                    .EnableRedirectionEndpointPassthrough();

                options.UseSystemNetHttp()
                    .SetProductInformation(typeof(IdentityExtensions).Assembly);

                options.UseWebProviders()
                    .AddDiscord(options =>
                    {
                        options.SetClientId($"{discordConfig.ClientId}")
                            .SetClientSecret($"{discordConfig.ClientSecret}")
                            .SetRedirectUri("callback/login/discord");
                    });
            })
            .AddServer(options =>
            {
                options.SetConfigurationEndpointUris("/.well-known/openid-configuration")
                    .SetJsonWebKeySetEndpointUris("/.well-known/jwks")
                    .SetAuthorizationEndpointUris("/authorize")
                    .SetEndSessionEndpointUris("/logout")
                    .SetTokenEndpointUris("/token")
                    .SetUserInfoEndpointUris("/userinfo");

                options.AllowAuthorizationCodeFlow()
                    .AllowRefreshTokenFlow();

                options.RegisterScopes(OpenIddictConstants.Scopes.Roles);

                if (identityConfig.IsEmpty && builder.Environment.IsDevelopment())
                {
                    options.AddDevelopmentEncryptionCertificate()
                        .AddDevelopmentSigningCertificate();
                }
                else
                {
                    options.AddSigningCertificate(identityConfig.LoadSigningCertificate())
                        .AddEncryptionCertificate(identityConfig.LoadEncryptionCertificate());
                }

                options.UseAspNetCore()
                    .EnableAuthorizationEndpointPassthrough();
            })
            .AddValidation(options =>
            {
                options.UseLocalServer();
                options.UseAspNetCore();
            });

        return builder;
    }
}
