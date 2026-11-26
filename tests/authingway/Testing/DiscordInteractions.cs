// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using NetCord;
using NetCord.Gateway;
using NetCord.JsonModels;
using NetCord.Rest;
using NetCord.Services.ApplicationCommands;
using NetCord.Services.ComponentInteractions;

namespace Naur.Authingway.Tests.Testing;

/// <summary>
/// Builds real NetCord interaction contexts for a member of the configured guild.
/// </summary>
public static class DiscordInteractions
{
    /// <summary>
    /// The configured guild identifier.
    /// </summary>
    public const ulong GuildId = 1;

    /// <summary>
    /// A role identifier mapped to the admin role.
    /// </summary>
    public const ulong AdminRoleId = 10;

    /// <summary>
    /// Creates a gateway client whose REST calls go to <paramref name="rest"/>.
    /// </summary>
    /// <param name="rest">The fake REST handler.</param>
    /// <returns>The client.</returns>
    public static GatewayClient CreateClient(FakeDiscordRest rest)
    {
        return new GatewayClient(
            new BotToken("MTIzNDU2Nzg5MDEyMzQ1Njc4.test.test"),
            new GatewayClientConfiguration { RestClientConfiguration = new RestClientConfiguration { RequestHandler = rest } });
    }

    /// <summary>
    /// Creates a slash command context.
    /// </summary>
    /// <param name="client">The gateway client.</param>
    /// <param name="name">The command name.</param>
    /// <param name="options">The string option values.</param>
    /// <returns>The context.</returns>
    public static ApplicationCommandContext Command(GatewayClient client, string name, params (string Name, ApplicationCommandOptionType Type, string Value)[] options)
    {
        var interaction = Interaction.CreateFromJson(Base(new JsonInteractionData
        {
            Name = name,
            Type = ApplicationCommandType.ChatInput,
            Options = [.. options.Select(o => new JsonApplicationCommandInteractionDataOption { Name = o.Name, Type = o.Type, Value = o.Value })]
        }, InteractionType.ApplicationCommand), client.Cache, client.Rest);

        return new ApplicationCommandContext((ApplicationCommandInteraction)interaction, client);
    }

    /// <summary>
    /// Creates a modal submission context.
    /// </summary>
    /// <param name="client">The gateway client.</param>
    /// <param name="customId">The modal custom identifier.</param>
    /// <param name="components">The submitted label components.</param>
    /// <returns>The context.</returns>
    public static ModalInteractionContext ModalSubmit(GatewayClient client, string customId, params JsonComponent[] components)
    {
        var interaction = Interaction.CreateFromJson(Base(new JsonInteractionData
        {
            CustomId = customId,
            Components = components
        }, InteractionType.Modal), client.Cache, client.Rest);

        return new ModalInteractionContext((ModalInteraction)interaction, client);
    }

    /// <summary>
    /// Creates a submitted text input wrapped in a label.
    /// </summary>
    /// <param name="customId">The input identifier.</param>
    /// <param name="value">The submitted value.</param>
    /// <returns>The component.</returns>
    public static JsonComponent Text(string customId, string value)
    {
        return new JsonLabelComponent { Type = ComponentType.Label, Component = new JsonTextInputComponent { Type = ComponentType.TextInput, CustomId = customId, Value = value } };
    }

    /// <summary>
    /// Creates a submitted checkbox group wrapped in a label.
    /// </summary>
    /// <param name="customId">The group identifier.</param>
    /// <param name="values">The checked values.</param>
    /// <returns>The component.</returns>
    public static JsonComponent Checked(string customId, params string[] values)
    {
        return new JsonLabelComponent { Type = ComponentType.Label, Component = new JsonCheckboxGroupComponent { Type = ComponentType.CheckboxGroup, CustomId = customId, CheckedValues = values } };
    }

    private static JsonInteraction Base(JsonInteractionData data, InteractionType type)
    {
        var user = new JsonUser { Id = 42, Username = "admin" };
        return new JsonInteraction
        {
            Id = 1,
            ApplicationId = 2,
            Token = "token",
            Type = type,
            GuildId = GuildId,
            Channel = new JsonChannel { Id = 100, Type = ChannelType.TextGuildChannel },
            User = user,
            GuildUser = new JsonGuildUser { User = user, RoleIds = [AdminRoleId] },
            Data = data,
            Entitlements = []
        };
    }
}
