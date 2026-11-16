// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration;
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
    /// The longest a migration may run before it is canceled, kept inside Discord's 15 minute window for editing a
    /// deferred response.
    /// </summary>
    public static readonly TimeSpan Timeout = TimeSpan.FromMinutes(10);

    /// <summary>
    /// Applies any pending database migrations.
    /// </summary>
    /// <returns>A task that represents the asynchronous command execution.</returns>
    /// <remarks>
    /// Invocations are acknowledged immediately and the acknowledgement is replaced with the outcome once the
    /// migration finishes or <see cref="Timeout"/> elapses.
    /// </remarks>
    [SlashCommand("migrate", "Applies pending database migrations.")]
    [RequireDiscordRoles<ApplicationCommandContext>(
        Roles = [DiscordRole.Admin, DiscordRole.Developer],
        Channels = [DiscordChannelGroup.Admin, DiscordChannelGroup.Developer])]
    public async Task MigrateAsync()
    {
        await RespondAsync(InteractionCallback.DeferredMessage());

        var executor = $"{Context.User.Username} ({Context.User.Id})";

        string content;

        try
        {
            using var timeout = new CancellationTokenSource(Timeout);

            content = await migrator.MigrateAsync(executor, timeout.Token) switch
            {
                MigrationOutcome.Applied => "Database migrations applied.",
                MigrationOutcome.UpToDate => "Database is already up to date.",
                MigrationOutcome.AlreadyRunning => "A migration is already running.",
                MigrationOutcome.Failed => "Database migrations failed; check the authingway logs.",
                _ => "The migration finished in an unrecognised state."
            };
        }
        catch (OperationCanceledException)
        {
            content = "Database migration timed out; check the authingway logs.";
        }

        await ModifyResponseAsync(message => message.Content = content);
    }
}
