// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Reflection;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Discord.Preconditions;
using Naur.Authingway.Identity.Commands;
using Naur.Authingway.Identity.Modals;
using NetCord;
using NetCord.Gateway;
using NetCord.JsonModels;
using NetCord.Services;
using NetCord.Services.ComponentInteractions;
using ModalPreconditions = Naur.Authingway.Discord.Preconditions.Modals;

namespace Naur.Authingway.Tests.Identity;

/// <summary>
/// Tests that client management is restricted to the configured guild and Admin/Developer roles.
/// </summary>
public class ClientAuthorizationTests
{
    /// <summary>
    /// Gets every client management command and modal.
    /// </summary>
    /// <returns>The module types.</returns>
    public static IEnumerable<Func<Type>> Handlers()
    {
        yield return () => typeof(CreateClientCommand);
        yield return () => typeof(ListClientsCommand);
        yield return () => typeof(ViewClientCommand);
        yield return () => typeof(EditClientCommand);
        yield return () => typeof(DeleteClientCommand);
        yield return () => typeof(CreateClientModal);
        yield return () => typeof(EditClientModal);
        yield return () => typeof(DeleteClientModal);
    }

    /// <summary>
    /// Verifies each module requires the configured guild and the Admin or Developer role.
    /// </summary>
    /// <param name="module">The module type.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [MethodDataSource(nameof(Handlers))]
    public async Task ModuleRequiresGuildAndManagementRoles(Type module)
    {
        var attributes = module.GetCustomAttributesData();
        var guild = attributes.Count(a => IsGeneric(a.AttributeType, typeof(AuthorizeGuildAttribute<>)));
        var roles = attributes.Where(a => IsGeneric(a.AttributeType, typeof(AuthorizeRolesAttribute<>))).ToList();

        await Assert.That(guild).IsEqualTo(1);
        await Assert.That(roles.Count).IsEqualTo(1);

        var requested = ((IEnumerable<CustomAttributeTypedArgument>)roles[0].ConstructorArguments[0].Value!)
            .Select(argument => (DiscordRole)argument.Value!)
            .Order();

        await Assert.That(requested.SequenceEqual([DiscordRole.Admin, DiscordRole.Developer])).IsTrue();
    }

    /// <summary>
    /// Verifies that NetCord discovers the modal submit handlers and both preconditions on them.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task ModalHandlersRegisterPreconditions()
    {
        var service = new ComponentInteractionService<ModalInteractionContext>();
        service.AddModule<CreateClientModal>();
        service.AddModule<EditClientModal>();
        service.AddModule<DeleteClientModal>();

        var interactions = service.GetComponentInteractions();

        await Assert.That(interactions.Count).IsEqualTo(3);

        foreach (var interaction in interactions.Values)
        {
            await Assert.That(interaction.Preconditions.Count).IsEqualTo(2);
        }
    }

    /// <summary>
    /// Verifies that the generic preconditions authorize modal submissions.
    /// </summary>
    /// <param name="guildId">The invoking guild.</param>
    /// <param name="roleId">The invoking member's role.</param>
    /// <param name="expected">Whether access should be granted.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments(1UL, 10UL, true)]
    [Arguments(1UL, 20UL, true)]
    [Arguments(1UL, 999UL, false)]
    [Arguments(2UL, 10UL, false)]
    public async Task AuthorizesModalSubmissions(ulong guildId, ulong roleId, bool expected)
    {
        using var services = new ServiceCollection()
            .AddSingleton(Options.Create(new DiscordConfig
            {
                GuildId = 1,
                Roles = { [DiscordRole.Admin] = 10, [DiscordRole.Developer] = 20 }
            }))
            .BuildServiceProvider();
        using var client = new GatewayClient(new BotToken("MTIzNDU2Nzg5MDEyMzQ1Njc4.test.test"));

        var user = new JsonUser { Id = 42, Username = "tester" };
        var interaction = (ModalInteraction)Interaction.CreateFromJson(new JsonInteraction
        {
            Type = InteractionType.Modal,
            GuildId = guildId,
            Channel = new JsonChannel { Id = 100, Type = ChannelType.TextGuildChannel },
            User = user,
            GuildUser = new JsonGuildUser { User = user, RoleIds = [roleId] },
            Data = new JsonInteractionData { CustomId = ClientModal.CreateCustomId, Components = [] },
            Entitlements = []
        }, client.Cache, client.Rest);
        var context = new ModalInteractionContext(interaction, client);

        PreconditionAttribute<ModalInteractionContext>[] preconditions =
        [
            new ModalPreconditions.AuthorizeGuildAttribute(),
            new ModalPreconditions.AuthorizeRolesAttribute([DiscordRole.Admin, DiscordRole.Developer])
        ];

        var permitted = true;

        foreach (var precondition in preconditions)
        {
            permitted &= await precondition.EnsureCanExecuteAsync(context, services) is PreconditionSuccessResult;
        }

        await Assert.That(permitted).IsEqualTo(expected);
    }

    private static bool IsGeneric(Type attributeType, Type definition)
    {
        for (var type = attributeType; type is not null; type = type.BaseType)
        {
            if (type.IsGenericType && type.GetGenericTypeDefinition() == definition)
            {
                return true;
            }
        }

        return false;
    }
}
