// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using NetCord;
using NetCord.Rest;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Data.Commands;

/// <summary>
/// Exposes the database migration command to Discord.
/// </summary>
/// <param name="authorizer">Decides whether an invocation is permitted. Cannot be null.</param>
/// <param name="scopeFactory">Creates the scope the migration runs in. Cannot be null.</param>
public class MigrationCommandModule(MigrationCommandAuthorizer authorizer, IServiceScopeFactory scopeFactory)
    : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// Applies any pending database migrations.
    /// </summary>
    /// <returns>A task that represents the asynchronous command execution.</returns>
    /// <remarks>
    /// Refusals are visible only to the invoking user. Permitted invocations are acknowledged immediately and the
    /// acknowledgement is replaced with the outcome once the migration finishes.
    /// </remarks>
    [SlashCommand("migrate", "Applies pending database migrations.")]
    public async Task MigrateAsync()
    {
        var channelId = Context.Interaction.Channel.Id;
        var roleIds = Context.Interaction.User is GuildUser guildUser ? guildUser.RoleIds : [];

        if (!authorizer.IsAuthorized(Context.Interaction.GuildId, channelId, roleIds))
        {
            await RespondAsync(InteractionCallback.Message(new InteractionMessageProperties
            {
                Content = "This command is not available here.",
                Flags = MessageFlags.Ephemeral
            }));

            return;
        }

        await RespondAsync(InteractionCallback.DeferredMessage());

        await using var scope = scopeFactory.CreateAsyncScope();

        var migrator = scope.ServiceProvider.GetRequiredService<DatabaseMigrator>();

        var outcome = await migrator.MigrateAsync(CancellationToken.None);

        var content = outcome switch
        {
            MigrationOutcome.Applied => "Database migrations applied.",
            MigrationOutcome.AlreadyRunning => "A migration is already running.",
            _ => "The migration finished in an unrecognised state."
        };

        await ModifyResponseAsync(message => message.Content = content);
    }
}
