// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Discord;
using NetCord.Rest;
using NetCord.Services.ApplicationCommands;

namespace Naur.Authingway.Data.Commands;

/// <summary>
/// Exposes the database migration command to Discord.
/// </summary>
/// <param name="migrator">Applies the migrations within the command's service scope. Cannot be null.</param>
public class MigrationCommandModule(DatabaseMigrator migrator) : ApplicationCommandModule<ApplicationCommandContext>
{
    /// <summary>
    /// The name of the channel and role restriction that guards the migration command.
    /// </summary>
    public const string Restriction = "Migration";

    /// <summary>
    /// Applies any pending database migrations.
    /// </summary>
    /// <returns>A task that represents the asynchronous command execution.</returns>
    /// <remarks>
    /// Invocations are acknowledged immediately and the acknowledgement is replaced with the outcome once the
    /// migration finishes.
    /// </remarks>
    [SlashCommand("migrate", "Applies pending database migrations.")]
    [RequireChannelRole<ApplicationCommandContext>(Restriction)]
    public async Task MigrateAsync()
    {
        await RespondAsync(InteractionCallback.DeferredMessage());

        var outcome = await migrator.MigrateAsync(CancellationToken.None);

        var content = outcome switch
        {
            MigrationOutcome.Applied => "Database migrations applied.",
            MigrationOutcome.AlreadyRunning => "A migration is already running.",
            MigrationOutcome.Failed => "Database migrations failed; check the authingway logs.",
            _ => "The migration finished in an unrecognised state."
        };

        await ModifyResponseAsync(message => message.Content = content);
    }
}
