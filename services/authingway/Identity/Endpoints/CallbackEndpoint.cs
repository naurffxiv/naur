// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Security.Claims;
using FastEndpoints;
using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.Authentication.Cookies;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using NetCord.Gateway;
using NetCord.Rest;
using static OpenIddict.Client.WebIntegration.OpenIddictClientWebIntegrationConstants;

namespace Naur.Authingway.Identity.Endpoints;

/// <summary>
/// The oauth callback endpoint.
/// </summary>
public class CallbackEndpoint : EndpointWithoutRequest<Results<ChallengeHttpResult, SignInHttpResult>>
{
    /// <inheritdoc/>
    public override void Configure()
    {
        Verbs(Http.GET, Http.POST);
        Routes("/callback/login/discord");
        AllowAnonymous();
        Description(b =>
        {
            b.ExcludeFromDescription();
        });
    }

    /// <inheritdoc/>
    public override async Task<Results<ChallengeHttpResult, SignInHttpResult>> ExecuteAsync(CancellationToken ct)
    {
        var result = await HttpContext.AuthenticateAsync(Providers.Discord);

        var identity = new ClaimsIdentity(
            authenticationType: "ExternalLogin",
            nameType: ClaimTypes.Name,
            roleType: ClaimTypes.Role);

        identity.AddClaim(new Claim(ClaimTypes.NameIdentifier, result.Principal!.FindFirst("id")!.Value));

        foreach (var roleId in await GetGuildRoleIdsAsync(result.Principal!.FindFirst("id")!.Value, ct))
        {
            identity.AddClaim(new Claim(ClaimTypes.Role, roleId));
        }

        var properties = new AuthenticationProperties
        {
            RedirectUri = result.Properties!.RedirectUri
        };

        return TypedResults.SignIn(new ClaimsPrincipal(identity), properties, CookieAuthenticationDefaults.AuthenticationScheme);
    }

    private async Task<IEnumerable<string>> GetGuildRoleIdsAsync(string userId, CancellationToken ct)
    {
        var guildId = Resolve<IOptions<DiscordConfig>>().Value.GuildId;
        var gateway = TryResolve<GatewayClient>();

        if (guildId == 0 || gateway is null || !ulong.TryParse(userId, out var id))
        {
            return [];
        }

        try
        {
            var member = await gateway.Rest.GetGuildUserAsync(guildId, id, cancellationToken: ct);
            return member.RoleIds.Select(r => r.ToString());
        }
        catch (RestException)
        {
            // Not a member of the guild: no roles.
            return [];
        }
    }
}
