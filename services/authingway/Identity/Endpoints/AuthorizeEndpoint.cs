// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Security.Claims;
using FastEndpoints;
using Microsoft.AspNetCore;
using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.Authentication.Cookies;
using Microsoft.AspNetCore.Http.Extensions;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.IdentityModel.Tokens;
using OpenIddict.Abstractions;
using OpenIddict.Server.AspNetCore;
using static OpenIddict.Client.WebIntegration.OpenIddictClientWebIntegrationConstants;

namespace Naur.Authingway.Identity.Endpoints;

/// <summary>
/// The oauth authorize endpoint.
/// </summary>
public class AuthorizeEndpoint : EndpointWithoutRequest<Results<ChallengeHttpResult, SignInHttpResult>>
{
    /// <inheritdoc/>
    public override void Configure()
    {
        Verbs(Http.GET, Http.POST);
        Routes("/authorize");
        AllowAnonymous();
        Description(b =>
        {
            b.ExcludeFromDescription();
        });
    }

    /// <inheritdoc/>
    public override async Task<Results<ChallengeHttpResult, SignInHttpResult>> ExecuteAsync(CancellationToken ct)
    {
        var principal = (await HttpContext.AuthenticateAsync(CookieAuthenticationDefaults.AuthenticationScheme)).Principal;

        if (principal is not { Identity.IsAuthenticated: true })
        {
            var properties = new AuthenticationProperties
            {
                RedirectUri = HttpContext.Request.GetEncodedUrl()
            };

            return TypedResults.Challenge(properties, [Providers.Discord]);
        }

        var request = HttpContext.GetOpenIddictServerRequest()
            ?? throw new InvalidOperationException("The OpenIddict request is missing.");

        var identifier = principal.FindFirst(ClaimTypes.NameIdentifier)!.Value;

        var identity = new ClaimsIdentity(
            authenticationType: TokenValidationParameters.DefaultAuthenticationType,
            nameType: OpenIddictConstants.Claims.Name,
            roleType: OpenIddictConstants.Claims.Role);

        identity.SetClaim(OpenIddictConstants.Claims.Subject, identifier)
            .SetClaim(OpenIddictConstants.Claims.Name, identifier)
            .SetClaim(OpenIddictConstants.Claims.PreferredUsername, identifier)
            .SetClaims(OpenIddictConstants.Claims.Role, [.. principal.FindAll(ClaimTypes.Role).Select(role => role.Value)]);

        identity.SetScopes(request.GetScopes());
        identity.SetDestinations(GetDestinations);

        return TypedResults.SignIn(new ClaimsPrincipal(identity), properties: null, OpenIddictServerAspNetCoreDefaults.AuthenticationScheme);
    }

    private static IEnumerable<string> GetDestinations(Claim claim)
    {
        return claim.Type switch
        {
            OpenIddictConstants.Claims.Role =>
                claim.Subject!.HasScope(OpenIddictConstants.Scopes.Roles)
                    ? [OpenIddictConstants.Destinations.AccessToken, OpenIddictConstants.Destinations.IdentityToken]
                    : [OpenIddictConstants.Destinations.AccessToken],

            "AspNet.Identity.SecurityStamp" => [],

            _ => [OpenIddictConstants.Destinations.AccessToken]
        };
    }
}
