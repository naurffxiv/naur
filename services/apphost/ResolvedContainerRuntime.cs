// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

#pragma warning disable ASPIRECONTAINERRUNTIME001
#pragma warning disable ASPIREPIPELINES003

using Aspire.Hosting.Publishing;

internal sealed class ResolvedContainerRuntime(IContainerRuntimeResolver resolver) : IContainerRuntime
{
    public string Name => "Docker";

    public async Task<bool> CheckIfRunningAsync(CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        return await runtime.CheckIfRunningAsync(cancellationToken).ConfigureAwait(false);
    }

    public async Task BuildImageAsync(string contextPath, string dockerfilePath, ContainerImageBuildOptions? options, Dictionary<string, string?> buildArguments, Dictionary<string, BuildImageSecretValue> buildSecrets, string? stage, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.BuildImageAsync(contextPath, dockerfilePath, options, buildArguments, buildSecrets, stage, cancellationToken).ConfigureAwait(false);
    }

    public async Task TagImageAsync(string localImageName, string targetImageName, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.TagImageAsync(localImageName, targetImageName, cancellationToken).ConfigureAwait(false);
    }

    public async Task RemoveImageAsync(string imageName, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.RemoveImageAsync(imageName, cancellationToken).ConfigureAwait(false);
    }

    public async Task PushImageAsync(IResource resource, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.PushImageAsync(resource, cancellationToken).ConfigureAwait(false);
    }

    public async Task LoginToRegistryAsync(string registryServer, string username, string password, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.LoginToRegistryAsync(registryServer, username, password, cancellationToken).ConfigureAwait(false);
    }

    public async Task ComposeUpAsync(ComposeOperationContext context, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.ComposeUpAsync(context, cancellationToken).ConfigureAwait(false);
    }

    public async Task ComposeDownAsync(ComposeOperationContext context, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        await runtime.ComposeDownAsync(context, cancellationToken).ConfigureAwait(false);
    }

    public async Task<IReadOnlyList<ComposeServiceInfo>?> ComposeListServicesAsync(ComposeOperationContext context, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        return await runtime.ComposeListServicesAsync(context, cancellationToken).ConfigureAwait(false);
    }

    public async Task<ContainerImageConfigInspectionResult> InspectImageConfigAsync(string imageName, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        return await runtime.InspectImageConfigAsync(imageName, cancellationToken).ConfigureAwait(false);
    }

    public async Task<ContainerImageManifestInspectionResult> InspectImageManifestAsync(string imageName, CancellationToken cancellationToken)
    {
        var runtime = await resolver.ResolveAsync(cancellationToken).ConfigureAwait(false);
        return await runtime.InspectImageManifestAsync(imageName, cancellationToken).ConfigureAwait(false);
    }
}
