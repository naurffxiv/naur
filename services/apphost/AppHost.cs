// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Aspire.Hosting.Docker.Resources.ComposeNodes;
using Aspire.Hosting.Yarp.Transforms;

const string TunnelNetwork = "naur-tunnel";

var builder = DistributedApplication.CreateBuilder(args);

var composeEnvironment = builder.AddDockerComposeEnvironment("env")
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

var authingway = builder.AddProject<Projects.Naur_Authingway>("authingway")
    .WithReference(authingwayDb)
    .WaitFor(authingwayDb)
    .WithEnvironment("Discord__Token", discordToken)
    .WithHttpHealthCheck("/health");

if (!builder.ExecutionContext.IsPublishMode)
{
    authingway.WithExternalHttpEndpoints();
}

if (builder.ExecutionContext.IsPublishMode)
{
    composeEnvironment.ConfigureComposeFile(file =>
    {
        file.AddNetwork(new Network { Name = TunnelNetwork, External = true });
        file.Services[$"{composeEnvironment.Resource.Name}-dashboard"].Ports = ["127.0.0.1:18888:18888"];
    });

    builder.AddYarp("ingress")
        .WithConfiguration(yarp =>
        {
            yarp.AddRoute(authingway)
                .WithTransformUseOriginalHostHeader();
        })
        .WithEnvironment("ASPNETCORE_FORWARDEDHEADERS_ENABLED", "true")
        .PublishAsDockerComposeService((_, service) =>
        {
            service.Networks.Add(TunnelNetwork);
        });
}

builder.Build().Run();
