// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

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

var discordToken = builder.AddParameter("DiscordToken", secret: true);
var discordGuildId = builder.AddParameter("DiscordGuildId", "0", publishValueAsDefault: true);
var discordMigrationChannelId = builder.AddParameter("DiscordMigrationChannelId", "0", publishValueAsDefault: true);
var discordDeveloperRoleId = builder.AddParameter("DiscordDeveloperRoleId", "0", publishValueAsDefault: true);

var authingway = builder.AddProject<Projects.Naur_Authingway>("authingway")
    .WithReference(authingwayDb)
    .WaitFor(authingwayDb)
    .WithEnvironment("Discord__Token", discordToken)
    .WithEnvironment("Discord__Restrictions__Migration__GuildId", discordGuildId)
    .WithEnvironment("Discord__Restrictions__Migration__ChannelId", discordMigrationChannelId)
    .WithEnvironment("Discord__Restrictions__Migration__RoleId", discordDeveloperRoleId)
    .WithHttpHealthCheck("/health");

if (builder.ExecutionContext.IsPublishMode)
{
    authingway.WithHttpEndpoint(targetPort: 8080)
        .WithEnvironment("OpenIddict__SigningCertificate", builder.AddParameter("SigningCertificate", secret: true))
        .WithEnvironment("OpenIddict__SigningCertificatePassword", builder.AddParameter("SigningCertificatePassword", secret: true))
        .WithEnvironment("OpenIddict__EncryptionCertificate", builder.AddParameter("EncryptionCertificate", secret: true))
        .WithEnvironment("OpenIddict__EncryptionCertificatePassword", builder.AddParameter("EncryptionCertificatePassword", secret: true));

    var cloudflareToken = builder.AddParameter("CloudflareToken", secret: true);

    builder.AddContainer("cloudflared", "cloudflare/cloudflared", "2026.10.7")
        .WithArgs("tunnel", "--no-autoupdate", "run")
        .WithEnvironment("TUNNEL_TOKEN", cloudflareToken)
        .WaitFor(authingway);
}

builder.Build().Run();
