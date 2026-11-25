// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Diagnostics;
using Aspire.Hosting.Pipelines;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

var builder = DistributedApplication.CreateBuilder(args);

builder.AddDockerComposeEnvironment("env")
    .WithDashboard(enabled: false)
    .WithSshDeploySupport();

var postgres = builder.AddPostgres("postgres")
    .WithPassword(builder.AddParameter("PostgresPassword", true))
    .WithLifetime(ContainerLifetime.Persistent)
    .WithDataVolume()
    .WithEndpoint("tcp", endpoint =>
    {
        endpoint.Port = 5432;
        endpoint.TargetPort = 5432;
    })
    .WithPgAdmin(pgAdmin =>
    {
        pgAdmin.WithLifetime(ContainerLifetime.Persistent);
    });

var authingwayDb = postgres.AddDatabase("authingwaydb", databaseName: "authingway");

var authingway = builder.AddDotnetProject("authingway", "../authingway")
    .WithReference(authingwayDb)
    .WaitFor(authingwayDb)
    .WithEnvironment("Discord__Token", builder.AddParameter("DiscordToken", secret: true))
    .WithEnvironment("Discord__ClientId", builder.AddParameter("DiscordClientId", secret: true))
    .WithEnvironment("Discord__ClientSecret", builder.AddParameter("DiscordClientSecret", secret: true))
    .WithHttpHealthCheck("/health");

var authingwayMigrations = authingway.AddEFMigrations("authingway-migrations")
    .RunDatabaseUpdateOnStart()
    .WithReference(authingwayDb)
    .WaitFor(authingwayDb)
    .PublishAsMigrationBundle(publishContainer: true, selfContained: true);

authingway.WaitForCompletion(authingwayMigrations);

if (builder.ExecutionContext.IsPublishMode)
{
    authingway.WithHttpEndpoint(targetPort: 8080)
        .WithEnvironment("Identity__SigningCertificate", builder.AddParameter("SigningCertificate", secret: true))
        .WithEnvironment("Identity__SigningCertificatePassword", builder.AddParameter("SigningCertificatePassword", secret: true))
        .WithEnvironment("Identity__EncryptionCertificate", builder.AddParameter("EncryptionCertificate", secret: true))
        .WithEnvironment("Identity__EncryptionCertificatePassword", builder.AddParameter("EncryptionCertificatePassword", secret: true));

    builder.AddContainer("cloudflared", "cloudflare/cloudflared", "2026.10.7")
        .WithArgs("tunnel", "--no-autoupdate", "run")
        .WithEnvironment("TUNNEL_TOKEN", builder.AddParameter("CloudflareToken", secret: true))
        .WaitFor(authingway);
}

builder.Pipeline.AddStep("prune-images-env", async context =>
{
    var configuration = context.Services.GetRequiredService<IConfiguration>();
    var target = $"{configuration["DockerSSH:SshUsername"]}@{configuration["DockerSSH:TargetHost"]}";

    await using var task = await context.ReportingStep.CreateTaskAsync("Pruning unused images", context.CancellationToken);

    var result = await Process.RunAndCaptureTextAsync("ssh", [target, "docker", "image", "prune", "-a", "-f"], context.CancellationToken);

    if (result.ExitStatus.ExitCode == 0)
    {
        await task.SucceedAsync(result.StandardOutput.Trim(), context.CancellationToken);
    }
    else
    {
        await task.WarnAsync($"docker image prune exited with {result.ExitStatus.ExitCode}: {result.StandardError.Trim()}", context.CancellationToken);
    }
}, dependsOn: "health-check-env", requiredBy: "deploy-docker-ssh-env");

builder.Build().Run();
