// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Identity.Modals;
using Naur.Authingway.Identity.Models;
using Naur.Authingway.Tests.Testing;
using NetCord.Rest;
using static OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Tests.Identity.Modals;

/// <summary>
/// Tests the client create, edit, and delete modals without Discord.
/// </summary>
public class ClientModalTests
{
    /// <summary>
    /// Verifies that the create modal starts from the default client settings.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task CreateModalStartsFromDefaults()
    {
        var modal = ClientModal.ForCreate();

        await Assert.That(modal.CustomId).IsEqualTo(ClientModal.CreateCustomId);
        await Assert.That(modal.Title).IsEqualTo("Create client");
        await Assert.That(Text(modal, ClientModal.DisplayNameInput).Value).IsNull();
        await Assert.That(Defaults(modal, ClientModal.GrantTypesInput)).IsEquivalentTo([GrantTypes.AuthorizationCode]);
        await Assert.That(Defaults(modal, ClientModal.ScopesInput)).IsEquivalentTo([Scopes.OpenId]);
        await Assert.That(Defaults(modal, ClientModal.OptionsInput)).IsEquivalentTo([ClientModal.RequirePkceOption]);
    }

    /// <summary>
    /// Verifies that the edit modal is prefilled with the client's current settings.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task EditModalIsPrefilled()
    {
        var settings = ClientSettings.Parse(
            "App", "https://a.example/cb\nhttps://b.example/cb", [GrantTypes.AuthorizationCode, GrantTypes.RefreshToken],
            [Scopes.Roles], requirePkce: false, isPublic: true).Settings!;

        var modal = ClientModal.ForEdit("client-edit:abc", settings);

        await Assert.That(modal.CustomId).IsEqualTo("client-edit:abc");
        await Assert.That(Text(modal, ClientModal.DisplayNameInput).Value).IsEqualTo("App");
        await Assert.That(Text(modal, ClientModal.RedirectUrisInput).Value).IsEqualTo("https://a.example/cb\nhttps://b.example/cb");
        await Assert.That(Defaults(modal, ClientModal.GrantTypesInput)).IsEquivalentTo([GrantTypes.AuthorizationCode, GrantTypes.RefreshToken]);
        await Assert.That(Defaults(modal, ClientModal.ScopesInput)).IsEquivalentTo([Scopes.Roles]);
        await Assert.That(Defaults(modal, ClientModal.OptionsInput)).IsEquivalentTo([ClientModal.PublicOption, ClientModal.RequirePkceOption]);
    }

    /// <summary>
    /// Verifies that submitted modal values become client settings.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ReadsSubmittedSettings()
    {
        using var client = DiscordInteractions.CreateClient(new FakeDiscordRest());
        var context = DiscordInteractions.ModalSubmit(client, ClientModal.CreateCustomId,
            DiscordInteractions.Text(ClientModal.DisplayNameInput, " App "),
            DiscordInteractions.Text(ClientModal.RedirectUrisInput, "https://app.example/cb"),
            DiscordInteractions.Checked(ClientModal.GrantTypesInput, GrantTypes.AuthorizationCode),
            DiscordInteractions.Checked(ClientModal.ScopesInput, Scopes.OpenId, Scopes.Roles),
            DiscordInteractions.Checked(ClientModal.OptionsInput, ClientModal.PublicOption));

        var settings = ClientModal.Read(context.Components).Settings!;

        await Assert.That(settings.DisplayName).IsEqualTo("App");
        await Assert.That(settings.Scopes.SetEquals([Scopes.OpenId, Scopes.Roles])).IsTrue();
        await Assert.That(settings.IsPublic).IsTrue();
        await Assert.That(settings.RequirePkce).IsTrue();
    }

    /// <summary>
    /// Verifies which client identifiers can be carried in a modal custom identifier.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <param name="allowed">Whether a custom identifier can be built.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("abc123", true)]
    [Arguments("", false)]
    [Arguments("a:b", false)]
    public async Task BuildsCustomIdsForClientIds(string clientId, bool allowed)
    {
        await Assert.That(ClientModal.TryCreateCustomId(ClientModal.EditCustomId, clientId, out var customId)).IsEqualTo(allowed);
        await Assert.That(customId).IsEqualTo($"{ClientModal.EditCustomId}:{clientId}");
    }

    /// <summary>
    /// Verifies that custom identifiers stay within Discord's 100 character limit.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task RejectsCustomIdsOverDiscordsLimit()
    {
        var longest = new string('a', 100 - ClientModal.EditCustomId.Length - 1);

        await Assert.That(ClientModal.TryCreateCustomId(ClientModal.EditCustomId, longest, out _)).IsTrue();
        await Assert.That(ClientModal.TryCreateCustomId(ClientModal.EditCustomId, longest + "a", out _)).IsFalse();
    }

    /// <summary>
    /// Verifies that the delete modal asks for the exact client identifier and shortens long names.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task DeleteModalAsksForTheClientId()
    {
        var modal = DeleteClientModal.Build("client-delete:abc", "abc", new string('n', 200));

        var label = (LabelProperties)modal.Components.Single();
        var input = (TextInputProperties)label.Component;
        await Assert.That(input.MinLength).IsEqualTo(3);
        await Assert.That(input.MaxLength).IsEqualTo(3);
        await Assert.That(input.Placeholder).IsEqualTo("abc");
        await Assert.That(label.Description!.Length).IsEqualTo(100);
        await Assert.That(label.Description).EndsWith("…");
    }

    /// <summary>
    /// Verifies that deletion is confirmed only by the exact client identifier.
    /// </summary>
    /// <param name="typed">The typed confirmation.</param>
    /// <param name="confirmed">Whether deletion is confirmed.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("abc", true)]
    [Arguments("  abc ", true)]
    [Arguments("ABC", false)]
    [Arguments("ab", false)]
    public async Task ConfirmsDeletionWithTheExactClientId(string typed, bool confirmed)
    {
        using var client = DiscordInteractions.CreateClient(new FakeDiscordRest());
        var context = DiscordInteractions.ModalSubmit(client, "client-delete:abc", DiscordInteractions.Text(DeleteClientModal.ConfirmationInput, typed));

        await Assert.That(DeleteClientModal.IsConfirmed(context.Components, "abc")).IsEqualTo(confirmed);
    }

    private static TextInputProperties Text(ModalProperties modal, string customId)
    {
        return modal.Components.OfType<LabelProperties>().Select(l => l.Component).OfType<TextInputProperties>().Single(t => t.CustomId == customId);
    }

    private static string[] Defaults(ModalProperties modal, string customId)
    {
        var group = modal.Components.OfType<LabelProperties>().Select(l => l.Component).OfType<CheckboxGroupProperties>().Single(g => g.CustomId == customId);
        return [.. group.Options.Where(o => o.Default).Select(o => o.Value)];
    }
}
