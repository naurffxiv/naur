// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Aspire.Hosting.Publishing;
using Aspire.Hosting.Yarp;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

var builder = DistributedApplication.CreateBuilder(args);

#pragma warning disable ASPIRECONTAINERRUNTIME001
builder.Services.AddSingleton(sp => sp.GetRequiredService<IContainerRuntimeResolver>().ResolveAsync().GetAwaiter().GetResult());
#pragma warning restore ASPIRECONTAINERRUNTIME001

builder.AddDockerComposeEnvironment("env")
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
    var ingress = builder.AddYarp("ingress")
        .WithConfiguration(yarp =>
        {
            yarp.AddRoute(authingway);
        });

    if (!builder.Configuration.GetValue<bool>("EnableHttps"))
    {
        ingress.WithHostPort(80)
            .WithExternalHttpEndpoints();
    }
    else
    {
        var domain = builder.AddParameter("Domain");
        var letsEncryptEmail = builder.AddParameter("LetsEncryptEmail");
        var letsEncryptVolume = "letsencrypt";

        var certbot = builder.AddContainer("certbot", "certbot/certbot")
            .WithVolume(letsEncryptVolume, "/etc/letsencrypt")
            .WithHttpEndpoint(port: 80, targetPort: 80)
            .WithExternalHttpEndpoints()
            .WithArgs(
                "certonly",
                "--standalone",
                "--non-interactive",
                "--agree-tos",
                "-v",
                "--keep-until-expiring",
                "--deploy-hook",
                "chmod -R 755 /etc/letsencrypt/live && chmod -R 755 /etc/letsencrypt/archive",
                "--email",
                letsEncryptEmail.Resource,
                "-d",
                domain.Resource);

        ingress.WaitForCompletion(certbot)
            .WithVolume(letsEncryptVolume, "/etc/letsencrypt", isReadOnly: true)
            .WithHostPort(80)
            .WithHttpsEndpoint(443)
            .WithExternalHttpEndpoints()
            .WithEnvironment(context =>
            {
                context.EnvironmentVariables["Kestrel__Certificates__Default__Path"] =
                    ReferenceExpression.Create($"/etc/letsencrypt/live/{domain}/fullchain.pem");
                context.EnvironmentVariables["Kestrel__Certificates__Default__KeyPath"] =
                    ReferenceExpression.Create($"/etc/letsencrypt/live/{domain}/privkey.pem");

                var yarpResource = (YarpResource)context.Resource;
                var httpEndpoint = yarpResource.GetEndpoint("http");
                var httpsEndpoint = yarpResource.GetEndpoint("https");

                context.EnvironmentVariables["ASPNETCORE_URLS"] =
                    ReferenceExpression.Create($"http://+:{httpEndpoint.Property(EndpointProperty.TargetPort)};https://+:{httpsEndpoint.Property(EndpointProperty.TargetPort)}");
            });
    }
}

builder.Build().Run();
