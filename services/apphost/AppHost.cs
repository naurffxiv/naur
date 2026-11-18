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

var authingway = builder.AddProject<Projects.Naur_Authingway>("authingway")
    .WithReference(authingwayDb)
    .WaitFor(authingwayDb)
    .WithEnvironment("Discord__Token", builder.AddParameter("DiscordToken", secret: true))
    .WithHttpHealthCheck("/health");

authingway.AddEFMigrations("authingway-migrations")
    .RunDatabaseUpdateOnStart()
    .WithReference(authingwayDb)
    .WaitFor(authingwayDb)
    .PublishAsMigrationBundle(publishContainer: true);

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

builder.Build().Run();
