// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.Extensions.DependencyInjection;
using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Identity.Commands;
using Naur.Authingway.Identity.Modals;
using Naur.Authingway.Identity.Models;
using Naur.Authingway.Identity.Services;
using Naur.Authingway.Tests.Testing;
using NetCord;
using NetCord.Services;
using NetCord.Services.ApplicationCommands;
using NetCord.Services.ComponentInteractions;
using OpenIddict.Abstractions;
using static OpenIddict.Abstractions.OpenIddictConstants;

namespace Naur.Authingway.Tests.Identity.Commands;

/// <summary>
/// Drives the client management slash commands and modals through NetCord against the real client store.
/// </summary>
/// <param name="store">The Postgres-backed client store.</param>
[NotInParallel(nameof(ClientCommandsTests))]
[ClassDataSource<ClientStore>(Shared = SharedType.PerTestSession)]
public sealed class ClientCommandsTests(ClientStore store) : IDisposable
{
    private static readonly Lazy<ApplicationCommandService<ApplicationCommandContext>> s_commands = new(() =>
    {
        var commands = new ApplicationCommandService<ApplicationCommandContext>();
        commands.AddModules(typeof(CreateClientCommand).Assembly);
        return commands;
    });

    private static readonly Lazy<ComponentInteractionService<ModalInteractionContext>> s_modals = new(() =>
    {
        var modals = new ComponentInteractionService<ModalInteractionContext>();
        modals.AddModules(typeof(CreateClientModal).Assembly);
        return modals;
    });

    private readonly FakeDiscordRest _rest = new();
    private readonly List<IServiceScope> _readScopes = [];
    private AsyncServiceScope _scope;

    private IServiceProvider Services => _scope.ServiceProvider;

    /// <inheritdoc/>
    public void Dispose()
    {
        _rest.Dispose();
    }

    /// <summary>
    /// Starts each test in its own service scope with no clients.
    /// </summary>
    /// <returns>A task representing the setup.</returns>
    [Before(Test)]
    public async Task StartWithNoClients()
    {
        store.SkipIfUnavailable();
        _scope = store.Services.CreateAsyncScope();

        await using var cleanup = store.Services.CreateAsyncScope();
        var manager = cleanup.ServiceProvider.GetRequiredService<IOpenIddictApplicationManager>();
        var existing = new List<object>();
        await foreach (var application in manager.ListAsync())
        {
            existing.Add(application);
        }

        foreach (var application in existing)
        {
            await manager.DeleteAsync(application);
        }
    }

    /// <summary>
    /// Disposes the test's service scope.
    /// </summary>
    /// <returns>A task representing the cleanup.</returns>
    [After(Test)]
    public async Task EndScope()
    {
        foreach (var scope in _readScopes)
        {
            scope.Dispose();
        }

        await _scope.DisposeAsync();
    }

    /// <summary>
    /// Verifies that creating a client opens the create modal.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task CreateCommandOpensTheCreateModal()
    {
        await CommandAsync("client-create");

        await Assert.That(_rest.Modal.GetProperty("custom_id").GetString()).IsEqualTo(ClientModal.CreateCustomId);
    }

    /// <summary>
    /// Verifies that submitting the create modal registers a confidential client and shows its secret once.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task CreateModalRegistersAClient()
    {
        await SubmitAsync(ClientModal.CreateCustomId, Form("My_App"));

        var clientId = ClientIdIn(_rest.MessageContent);
        await Assert.That(_rest.MessageContent).StartsWith($"Created **My\\_App**.\nClient ID: `{clientId}`\nClient secret: `");
        var stored = await Clients().FindAsync(clientId);
        await Assert.That(stored!.ClientType).IsEqualTo(ClientTypes.Confidential);
        await Assert.That(stored.Settings.DisplayName).IsEqualTo("My_App");
    }

    /// <summary>
    /// Verifies that public clients are created without a secret.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task CreateModalRegistersAPublicClient()
    {
        await SubmitAsync(ClientModal.CreateCustomId, Form("Spa", ClientModal.PublicOption));

        await Assert.That(_rest.MessageContent).EndsWith("\nThis is a public client, so it has no secret.");
        var stored = await Clients().FindAsync(ClientIdIn(_rest.MessageContent));
        await Assert.That(stored!.ClientType).IsEqualTo(ClientTypes.Public);
    }

    /// <summary>
    /// Verifies that invalid input is reported and nothing is saved.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task CreateModalRejectsInvalidInput()
    {
        await SubmitAsync(ClientModal.CreateCustomId,
            DiscordInteractions.Text(ClientModal.DisplayNameInput, "App"),
            DiscordInteractions.Text(ClientModal.RedirectUrisInput, "ftp://app.example/cb"),
            DiscordInteractions.Checked(ClientModal.GrantTypesInput, GrantTypes.AuthorizationCode));

        await Assert.That(_rest.MessageContent).IsEqualTo(
            "The client was not saved:\n- `ftp://app.example/cb` is not an absolute HTTPS URI, or an HTTP URI for localhost or 127.0.0.1, without a fragment.");
        await Assert.That((await Clients().ListAsync(1)).Total).IsEqualTo(0);
    }

    /// <summary>
    /// Verifies the view command for existing and missing clients.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ViewCommandDescribesTheClient()
    {
        var created = await CreateAsync("Viewer");

        await CommandAsync("client-view", ("client_id", created.ClientId));
        await Assert.That(_rest.MessageContent).StartsWith($"**Viewer**\nClient ID: `{created.ClientId}`\nClient type: confidential\n");
        await CommandAsync("client-view", ("client_id", "nope"));
        await Assert.That(_rest.MessageContent).IsEqualTo("No client with ID `nope` exists.");
    }

    /// <summary>
    /// Verifies the list command with no clients, a page of clients, and an empty page.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ListCommandPagesThroughClients()
    {
        await CommandAsync("client-list");
        await Assert.That(_rest.MessageContent).IsEqualTo("No clients are registered. Use `/client-create` to add one.");

        var created = await CreateAsync("*Starred*");
        await CommandAsync("client-list");
        await Assert.That(_rest.MessageContent).IsEqualTo($"**Clients** (page 1 of 1, 1 total)\n- \\*Starred\\*: `{created.ClientId}`\n");
        await CommandAsync("client-list", ("page", "2"));
        await Assert.That(_rest.MessageContent).IsEqualTo("Page 2 is empty. There are 1 page(s).");
    }

    /// <summary>
    /// Verifies that editing opens a prefilled modal and that saving keeps the secret.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task EditCommandAndModalUpdateTheClient()
    {
        var created = await CreateAsync("Before");

        await CommandAsync("client-edit", ("client_id", created.ClientId));
        await Assert.That(_rest.Modal.GetProperty("custom_id").GetString()).IsEqualTo($"client-edit:{created.ClientId}");
        await SubmitAsync($"client-edit:{created.ClientId}", Form("After"));
        await Assert.That(_rest.MessageContent).IsEqualTo($"Updated **After** (`{created.ClientId}`). The client secret is unchanged.");
        await Assert.That((await Clients().FindAsync(created.ClientId))!.Settings.DisplayName).IsEqualTo("After");
    }

    /// <summary>
    /// Verifies the messages when a client switches between public and confidential.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task EditModalReportsSecretChanges()
    {
        var created = await CreateAsync("Switch");

        await SubmitAsync($"client-edit:{created.ClientId}", Form("Switch", ClientModal.PublicOption));
        await Assert.That(_rest.MessageContent).IsEqualTo($"Updated **Switch** (`{created.ClientId}`). This is a public client, so it has no secret.");
        await SubmitAsync($"client-edit:{created.ClientId}", Form("Switch"));
        await Assert.That(_rest.MessageContent).StartsWith(
            $"Updated **Switch** (`{created.ClientId}`).\nThe client is now confidential.\nClient secret: `");
    }

    /// <summary>
    /// Verifies the edit and delete commands for missing clients.
    /// </summary>
    /// <param name="command">The command.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments("client-edit")]
    [Arguments("client-delete")]
    public async Task EditAndDeleteReportMissingClients(string command)
    {
        await CommandAsync(command, ("client_id", "nope"));

        await Assert.That(_rest.MessageContent).IsEqualTo("No client with ID `nope` exists.");
    }

    /// <summary>
    /// Verifies the edit modal for a client deleted after the modal was opened.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task EditModalReportsMissingClients()
    {
        await SubmitAsync("client-edit:nope", Form("App"));

        await Assert.That(_rest.MessageContent).IsEqualTo("No client with ID `nope` exists.");
    }

    /// <summary>
    /// Verifies that deletion needs the client identifier typed back.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task DeleteCommandAndModalRemoveTheClient()
    {
        var created = await CreateAsync("Doomed");
        var customId = $"client-delete:{created.ClientId}";

        await CommandAsync("client-delete", ("client_id", created.ClientId));
        await Assert.That(_rest.Modal.GetProperty("custom_id").GetString()).IsEqualTo(customId);
        await SubmitAsync(customId, DiscordInteractions.Text(DeleteClientModal.ConfirmationInput, "wrong"));
        await Assert.That(_rest.MessageContent).IsEqualTo("The client ID did not match. Nothing was deleted.");
        await Assert.That(await Clients().FindAsync(created.ClientId)).IsNotNull();
        await SubmitAsync(customId, DiscordInteractions.Text(DeleteClientModal.ConfirmationInput, created.ClientId));
        await Assert.That(_rest.MessageContent).IsEqualTo($"Deleted client `{created.ClientId}`.");
        await Assert.That(await Clients().FindAsync(created.ClientId)).IsNull();
        await SubmitAsync(customId, DiscordInteractions.Text(DeleteClientModal.ConfirmationInput, created.ClientId));
        await Assert.That(_rest.MessageContent).IsEqualTo($"No client with ID `{created.ClientId}` exists.");
    }

    /// <summary>
    /// Verifies that members without an allowed role cannot run the commands.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task CommandsRequireAnAllowedRole()
    {
        var config = store.DiscordConfig;
        var original = config.Roles[DiscordRole.Admin];
        config.Roles[DiscordRole.Admin] = 999;
        try
        {
            var result = await CommandAsync("client-create");

            await Assert.That(result is IFailResult).IsTrue();
            await Assert.That(_rest.Callbacks).IsEmpty();
        }
        finally
        {
            config.Roles[DiscordRole.Admin] = original;
        }
    }

    private ClientRegistrationService Clients()
    {
        // A fresh scope, so reads see what the command saved rather than a cached entity.
        var scope = store.Services.CreateScope();
        _readScopes.Add(scope);
        return scope.ServiceProvider.GetRequiredService<ClientRegistrationService>();
    }

    private Task<ClientCredentials> CreateAsync(string displayName)
    {
        var settings = ClientSettings.Parse(displayName, "https://app.example/cb", [GrantTypes.AuthorizationCode], [Scopes.OpenId], requirePkce: true).Settings!;
        return Clients().CreateAsync(settings);
    }

    private static NetCord.JsonModels.JsonComponent[] Form(string displayName, params string[] options)
    {
        return
        [
            DiscordInteractions.Text(ClientModal.DisplayNameInput, displayName),
            DiscordInteractions.Text(ClientModal.RedirectUrisInput, "https://app.example/cb"),
            DiscordInteractions.Checked(ClientModal.GrantTypesInput, GrantTypes.AuthorizationCode),
            DiscordInteractions.Checked(ClientModal.ScopesInput, Scopes.OpenId),
            DiscordInteractions.Checked(ClientModal.OptionsInput, [ClientModal.RequirePkceOption, .. options])
        ];
    }

    private static string ClientIdIn(string message)
    {
        const string Marker = "Client ID: `";
        var start = message.IndexOf(Marker, StringComparison.Ordinal) + Marker.Length;
        return message[start..message.IndexOf('`', start)];
    }

    private async Task<IExecutionResult> CommandAsync(string name, params (string Name, string Value)[] options)
    {
        using var client = DiscordInteractions.CreateClient(_rest);
        var context = DiscordInteractions.Command(client, name,
            [.. options.Select(o => (o.Name, o.Name == "page" ? ApplicationCommandOptionType.Integer : ApplicationCommandOptionType.String, o.Value))]);

        var result = await s_commands.Value.ExecuteAsync(context, Services);

        if (result is IExceptionResult exception)
        {
            throw new InvalidOperationException("Command failed", exception.Exception);
        }

        return result;
    }

    private async Task SubmitAsync(string customId, params NetCord.JsonModels.JsonComponent[] components)
    {
        using var client = DiscordInteractions.CreateClient(_rest);

        var result = await s_modals.Value.ExecuteAsync(DiscordInteractions.ModalSubmit(client, customId, components), Services);

        if (result is IExceptionResult exception)
        {
            throw new InvalidOperationException("Modal failed", exception.Exception);
        }

        await Assert.That(result is IFailResult).IsFalse();
    }
}
