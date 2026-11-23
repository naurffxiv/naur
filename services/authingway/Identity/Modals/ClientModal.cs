// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Discord.Extensions;
using Naur.Authingway.Identity.Models;
using NetCord;
using NetCord.Rest;
using Constants = OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Identity.Modals;

/// <summary>
/// The shared form used to create and edit clients.
/// </summary>
public static class ClientModal
{
    /// <summary>
    /// The custom identifier of the create form.
    /// </summary>
    public const string CreateCustomId = "client-create";

    /// <summary>
    /// The custom identifier prefix of the edit form; the client identifier follows it.
    /// </summary>
    public const string EditCustomId = "client-edit";

    /// <summary>
    /// The display name input.
    /// </summary>
    public const string DisplayNameInput = "display_name";

    /// <summary>
    /// The redirect URIs input.
    /// </summary>
    public const string RedirectUrisInput = "redirect_uris";

    /// <summary>
    /// The grant types checkbox group.
    /// </summary>
    public const string GrantTypesInput = "grant_types";

    /// <summary>
    /// The scopes checkbox group.
    /// </summary>
    public const string ScopesInput = "scopes";

    /// <summary>
    /// The options checkbox group.
    /// </summary>
    public const string OptionsInput = "options";

    /// <summary>
    /// The public client option.
    /// </summary>
    public const string PublicOption = "public";

    /// <summary>
    /// The PKCE option.
    /// </summary>
    public const string RequirePkceOption = "require_pkce";

    private const int MaxCustomIdLength = 100;

    private static readonly Dictionary<string, (string Label, string Description)> s_grantTypeOptions = new()
    {
        [Constants.GrantTypes.AuthorizationCode] = ("Authorization code", "Required. Lets users sign in through Discord."),
        [Constants.GrantTypes.RefreshToken] = ("Refresh token", "Lets the client renew tokens using offline_access.")
    };

    private static readonly Dictionary<string, (string Label, string Description)> s_scopeOptions = new()
    {
        [Constants.Scopes.OpenId] = ("openid", "OpenID Connect sign-in with ID tokens."),
        [Constants.Scopes.Roles] = ("roles", "Includes the user's roles in issued tokens.")
    };

    /// <summary>
    /// Builds a custom identifier carrying a client identifier, when it can be encoded safely.
    /// </summary>
    /// <param name="prefix">The modal custom identifier prefix.</param>
    /// <param name="clientId">The client identifier.</param>
    /// <param name="customId">The custom identifier.</param>
    /// <returns>True when the client identifier can be encoded.</returns>
    public static bool TryCreateCustomId(string prefix, string clientId, out string customId)
    {
        customId = $"{prefix}:{clientId}";

        return !string.IsNullOrEmpty(clientId)
            && !clientId.Contains(':', StringComparison.Ordinal)
            && customId.Length <= MaxCustomIdLength;
    }

    /// <summary>
    /// Builds the create form with default values.
    /// </summary>
    /// <returns>The modal.</returns>
    public static ModalProperties ForCreate()
    {
        return Build(CreateCustomId, "Create client", ClientSettings.Default);
    }

    /// <summary>
    /// Builds the edit form prefilled with a client's current settings.
    /// </summary>
    /// <param name="customId">The custom identifier carrying the client identifier.</param>
    /// <param name="settings">The current settings.</param>
    /// <returns>The modal.</returns>
    public static ModalProperties ForEdit(string customId, ClientSettings settings)
    {
        return Build(customId, "Edit client", settings);
    }

    /// <summary>
    /// Reads and validates a submitted form.
    /// </summary>
    /// <param name="components">The submitted components.</param>
    /// <returns>The settings, or the validation errors.</returns>
    public static ClientSettingsResult Read(IEnumerable<IModalComponent> components)
    {
        ArgumentNullException.ThrowIfNull(components);

        var submitted = components.ToList();
        var options = submitted.GetCheckedValues(OptionsInput);

        return ClientSettings.Parse(
            submitted.GetText(DisplayNameInput),
            submitted.GetText(RedirectUrisInput),
            submitted.GetCheckedValues(GrantTypesInput),
            submitted.GetCheckedValues(ScopesInput),
            options.Contains(RequirePkceOption),
            options.Contains(PublicOption));
    }

    private static ModalProperties Build(string customId, string title, ClientSettings settings)
    {
        var displayName = new TextInputProperties(DisplayNameInput, TextInputStyle.Short)
            .WithMaxLength(ClientSettings.MaxDisplayNameLength)
            .WithRequired(true)
            .WithPlaceholder("My application");

        if (settings.DisplayName.Length > 0)
        {
            displayName.WithValue(settings.DisplayName);
        }

        var redirectUris = new TextInputProperties(RedirectUrisInput, TextInputStyle.Paragraph)
            .WithMaxLength(4000)
            .WithRequired(true)
            .WithPlaceholder("https://example.com/callback");

        if (settings.RedirectUris.Count > 0)
        {
            redirectUris.WithValue(string.Join('\n', settings.RedirectUris.Select(uri => uri.AbsoluteUri)));
        }

        var grantTypes = new CheckboxGroupProperties(GrantTypesInput, ClientSettings.SupportedGrantTypes
                .Select(grant => Option(s_grantTypeOptions[grant], grant, settings.GrantTypes.Contains(grant))))
            .WithMinValues(1)
            .WithRequired(true);

        var scopes = new CheckboxGroupProperties(ScopesInput, ClientSettings.SupportedScopes
                .Select(scope => Option(s_scopeOptions[scope], scope, settings.Scopes.Contains(scope))))
            .WithMinValues(0)
            .WithRequired(false);

        var options = new CheckboxGroupProperties(OptionsInput,
            [
                Option(("Public client", "No client secret, for apps that cannot keep one (SPAs, native apps). Always requires PKCE."), PublicOption, settings.IsPublic),
                Option(("Require PKCE", "Recommended. Clients must send a code challenge."), RequirePkceOption, settings.RequirePkce)
            ])
            .WithMinValues(0)
            .WithRequired(false);

        return new ModalProperties(customId, title,
        [
            new LabelProperties("Display name", displayName),
            new LabelProperties("Redirect URIs", redirectUris)
                .WithDescription("Absolute http(s) URIs, one per line."),
            new LabelProperties("Grant types", grantTypes),
            new LabelProperties("Scopes", scopes),
            new LabelProperties("Options", options)
        ]);
    }

    private static CheckboxGroupOptionProperties Option((string Label, string Description) option, string value, bool selected)
    {
        return new CheckboxGroupOptionProperties(option.Label, value)
            .WithDescription(option.Description)
            .WithDefault(selected);
    }
}
