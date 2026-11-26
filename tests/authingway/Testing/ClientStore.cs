// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Diagnostics;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Data;
using Naur.Authingway.Data.Entities;
using Naur.Authingway.Identity.Services;
using Npgsql;
using TUnit.Core.Interfaces;

namespace Naur.Authingway.Tests.Testing;

/// <summary>
/// Runs the real OpenIddict client store against a throwaway Postgres container,
/// without the rest of the application. Tests using it are skipped when Docker is unavailable.
/// </summary>
public sealed class ClientStore : IAsyncInitializer, IAsyncDisposable
{
    private string? _container;
    private ServiceProvider? _services;

    /// <summary>
    /// Gets the Discord configuration used by the authorization preconditions.
    /// </summary>
    public DiscordConfig DiscordConfig { get; } = new()
    {
        GuildId = DiscordInteractions.GuildId,
        Roles = { [DiscordRole.Admin] = DiscordInteractions.AdminRoleId }
    };

    /// <summary>
    /// Gets the services, skipping the test when Docker is unavailable.
    /// </summary>
    public IServiceProvider Services => _services ?? throw new InvalidOperationException("Not initialized.");

    /// <summary>
    /// Gets a value indicating whether Postgres is available.
    /// </summary>
    public bool Available => _services is not null;

    /// <inheritdoc/>
    public async Task InitializeAsync()
    {
        var (started, output) = await RunAsync("docker", "run", "-d", "--rm", "-e", "POSTGRES_PASSWORD=test", "-p", "127.0.0.1::5432",
            "--tmpfs", "/var/lib/postgresql/data", "postgres:16", "-c", "fsync=off", "-c", "synchronous_commit=off", "-c", "full_page_writes=off");
        if (!started)
        {
            return;
        }

        _container = output.Trim();
        var (_, port) = await RunAsync("docker", "port", _container, "5432/tcp");
        var connectionString = $"Host=127.0.0.1;Port={port.Trim().Split(':')[^1]};Username=postgres;Password=test;Database=postgres";

        var services = new ServiceCollection()
            .AddLogging()
            .AddSingleton(Options.Create(DiscordConfig))
            .AddDbContext<AppDbContext>(options => options
                .UseNpgsql(connectionString)
                .UseOpenIddict<Application, Authorization, Scope, Token, Guid>())
            .AddScoped<ClientRegistrationService>();
        services.AddOpenIddict().AddCore(options => options.UseEntityFrameworkCore()
            .UseDbContext<AppDbContext>()
            .ReplaceDefaultEntities<Application, Authorization, Scope, Token, Guid>());
        _services = services.BuildServiceProvider();

        await WaitForPostgresAsync(connectionString);
        await using var scope = _services.CreateAsyncScope();
        await scope.ServiceProvider.GetRequiredService<AppDbContext>().Database.MigrateAsync();
    }

    /// <summary>
    /// Skips the calling test when Docker is unavailable.
    /// </summary>
    public void SkipIfUnavailable()
    {
        Skip.When(!Available, "Docker is not available");
    }

    /// <inheritdoc/>
    public async ValueTask DisposeAsync()
    {
        if (_services is not null)
        {
            await _services.DisposeAsync();
        }

        if (_container is not null)
        {
            await RunAsync("docker", "stop", _container);
        }
    }

    private static async Task WaitForPostgresAsync(string connectionString)
    {
        var deadline = DateTime.UtcNow.AddSeconds(60);
        while (true)
        {
            try
            {
                await using var connection = new NpgsqlConnection(connectionString);
                await connection.OpenAsync();
                return;
            }
            catch (NpgsqlException) when (DateTime.UtcNow < deadline)
            {
                await Task.Delay(500);
            }
        }
    }

    private static async Task<(bool Succeeded, string Output)> RunAsync(string fileName, params string[] arguments)
    {
        try
        {
            using var process = Process.Start(new ProcessStartInfo(fileName, arguments)
            {
                RedirectStandardOutput = true,
                RedirectStandardError = true
            })!;
            var output = await process.StandardOutput.ReadToEndAsync();
            await process.WaitForExitAsync();
            return (process.ExitCode == 0, output);
        }
        catch (System.ComponentModel.Win32Exception)
        {
            return (false, string.Empty);
        }
    }
}
