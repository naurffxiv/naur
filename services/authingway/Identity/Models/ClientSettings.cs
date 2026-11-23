// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using OpenIddict.Abstractions;
using Constants = OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Identity.Models;

/// <summary>
/// The client registration settings that can be managed from Discord, limited to the
/// flows and scopes the authorization server supports.
/// </summary>
public sealed record ClientSettings
{
    /// <summary>
    /// The maximum length of a client display name.
    /// </summary>
    public const int MaxDisplayNameLength = 100;

    /// <summary>
    /// The maximum number of redirect URIs a client may register.
    /// </summary>
    public const int MaxRedirectUris = 10;

    /// <summary>
    /// Gets the grant types the authorization server supports, in display order.
    /// </summary>
    public static IReadOnlyList<string> SupportedGrantTypes { get; } =
        [Constants.GrantTypes.AuthorizationCode, Constants.GrantTypes.RefreshToken];

    /// <summary>
    /// Gets the scopes the authorization server supports, in display order.
    /// </summary>
    public static IReadOnlyList<string> SupportedScopes { get; } =
        [Constants.Scopes.OpenId, Constants.Scopes.Roles];

    /// <summary>
    /// Gets the default settings offered when creating a client.
    /// </summary>
    public static ClientSettings Default { get; } = new()
    {
        DisplayName = string.Empty,
        RedirectUris = [],
        GrantTypes = new HashSet<string> { Constants.GrantTypes.AuthorizationCode },
        Scopes = new HashSet<string> { Constants.Scopes.OpenId },
        RequirePkce = true,
        IsPublic = false
    };

    private static readonly string[] s_managedPermissions =
    [
        Constants.Permissions.Endpoints.Authorization,
        Constants.Permissions.Endpoints.Token,
        Constants.Permissions.ResponseTypes.Code,
        .. SupportedGrantTypes.Select(grant => Constants.Permissions.Prefixes.GrantType + grant),
        .. SupportedScopes.Select(scope => Constants.Permissions.Prefixes.Scope + scope)
    ];

    /// <summary>
    /// Gets the client display name.
    /// </summary>
    public required string DisplayName { get; init; }

    /// <summary>
    /// Gets the absolute redirect URIs registered for the client.
    /// </summary>
    public required IReadOnlyList<Uri> RedirectUris { get; init; }

    /// <summary>
    /// Gets the permitted grant types.
    /// </summary>
    public required IReadOnlySet<string> GrantTypes { get; init; }

    /// <summary>
    /// Gets the permitted scopes.
    /// </summary>
    public required IReadOnlySet<string> Scopes { get; init; }

    /// <summary>
    /// Gets a value indicating whether the client must use PKCE.
    /// </summary>
    public required bool RequirePkce { get; init; }

    /// <summary>
    /// Gets a value indicating whether the client is public, meaning it has no secret.
    /// </summary>
    public required bool IsPublic { get; init; }

    /// <summary>
    /// Validates raw form input and converts it into client settings.
    /// </summary>
    /// <param name="displayName">The display name.</param>
    /// <param name="redirectUris">The redirect URIs, one per line.</param>
    /// <param name="grantTypes">The selected grant types.</param>
    /// <param name="scopes">The selected scopes.</param>
    /// <param name="requirePkce">Whether PKCE is required. Always true for public clients.</param>
    /// <param name="isPublic">Whether the client is public.</param>
    /// <returns>The settings, or the validation errors when the input is invalid.</returns>
    public static ClientSettingsResult Parse(
        string? displayName,
        string? redirectUris,
        IEnumerable<string> grantTypes,
        IEnumerable<string> scopes,
        bool requirePkce,
        bool isPublic = false)
    {
        ArgumentNullException.ThrowIfNull(grantTypes);
        ArgumentNullException.ThrowIfNull(scopes);

        var errors = new List<string>();

        var name = displayName?.Trim() ?? string.Empty;

        if (name.Length == 0)
        {
            errors.Add("A display name is required.");
        }
        else if (name.Length > MaxDisplayNameLength)
        {
            errors.Add($"The display name must be at most {MaxDisplayNameLength} characters.");
        }

        var uris = new List<Uri>();
        var invalidUris = false;

        foreach (var line in (redirectUris ?? string.Empty).Split('\n', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries))
        {
            if (!Uri.TryCreate(line, UriKind.Absolute, out var uri)
                || !IsValidRedirectUri(uri))
            {
                invalidUris = true;
                errors.Add($"`{line}` is not an absolute HTTPS URI, or an HTTP URI for localhost or 127.0.0.1, without a fragment.");
            }
            else if (!uris.Contains(uri))
            {
                uris.Add(uri);
            }
        }

        if (uris.Count == 0 && !invalidUris)
        {
            errors.Add("At least one redirect URI is required.");
        }
        else if (uris.Count > MaxRedirectUris)
        {
            errors.Add($"At most {MaxRedirectUris} redirect URIs are allowed.");
        }

        var grants = grantTypes.ToHashSet(StringComparer.Ordinal);
        var scopeSet = scopes.ToHashSet(StringComparer.Ordinal);

        if (grants.Except(SupportedGrantTypes).Any() || scopeSet.Except(SupportedScopes).Any())
        {
            errors.Add("Only supported grant types and scopes can be selected.");
        }

        if (!grants.Contains(Constants.GrantTypes.AuthorizationCode))
        {
            errors.Add("The authorization code grant is required.");
        }

        return errors.Count > 0
            ? new ClientSettingsResult(null, errors)
            : new ClientSettingsResult(new ClientSettings
            {
                DisplayName = name,
                RedirectUris = uris,
                GrantTypes = grants,
                Scopes = scopeSet,
                RequirePkce = requirePkce || isPublic,
                IsPublic = isPublic
            }, []);
    }

    private static bool IsValidRedirectUri(Uri uri)
    {
        if (!string.IsNullOrEmpty(uri.Fragment))
        {
            return false;
        }

        if (uri.Scheme == Uri.UriSchemeHttps)
        {
            return true;
        }

        return uri.Scheme == Uri.UriSchemeHttp
            && (uri.Host.Equals("localhost", StringComparison.OrdinalIgnoreCase)
                || uri.Host.Equals("127.0.0.1", StringComparison.Ordinal));
    }

    /// <summary>
    /// Reads the managed settings from an application descriptor.
    /// </summary>
    /// <param name="descriptor">The descriptor populated from a stored application.</param>
    /// <returns>The managed settings of the application.</returns>
    public static ClientSettings FromDescriptor(OpenIddictApplicationDescriptor descriptor)
    {
        ArgumentNullException.ThrowIfNull(descriptor);

        return new ClientSettings
        {
            DisplayName = descriptor.DisplayName ?? string.Empty,
            RedirectUris = [.. descriptor.RedirectUris],
            GrantTypes = SupportedGrantTypes
                .Where(grant => descriptor.Permissions.Contains(Constants.Permissions.Prefixes.GrantType + grant))
                .ToHashSet(StringComparer.Ordinal),
            Scopes = SupportedScopes
                .Where(scope => descriptor.Permissions.Contains(Constants.Permissions.Prefixes.Scope + scope))
                .ToHashSet(StringComparer.Ordinal),
            RequirePkce = descriptor.Requirements.Contains(Constants.Requirements.Features.ProofKeyForCodeExchange),
            IsPublic = string.Equals(descriptor.ClientType, Constants.ClientTypes.Public, StringComparison.Ordinal)
        };
    }

    /// <summary>
    /// Applies these settings to an application descriptor, replacing only the values managed here.
    /// The client secret is not managed here.
    /// </summary>
    /// <param name="descriptor">The descriptor to update.</param>
    public void ApplyTo(OpenIddictApplicationDescriptor descriptor)
    {
        ArgumentNullException.ThrowIfNull(descriptor);

        descriptor.DisplayName = DisplayName;
        descriptor.ClientType = IsPublic ? Constants.ClientTypes.Public : Constants.ClientTypes.Confidential;

        descriptor.RedirectUris.Clear();
        descriptor.RedirectUris.UnionWith(RedirectUris);

        descriptor.Permissions.ExceptWith(s_managedPermissions);
        descriptor.Permissions.UnionWith(
        [
            Constants.Permissions.Endpoints.Authorization,
            Constants.Permissions.Endpoints.Token,
            Constants.Permissions.ResponseTypes.Code,
            .. GrantTypes.Select(grant => Constants.Permissions.Prefixes.GrantType + grant),
            .. Scopes.Select(scope => Constants.Permissions.Prefixes.Scope + scope)
        ]);

        if (RequirePkce)
        {
            descriptor.Requirements.Add(Constants.Requirements.Features.ProofKeyForCodeExchange);
        }
        else
        {
            descriptor.Requirements.Remove(Constants.Requirements.Features.ProofKeyForCodeExchange);
        }
    }
}
