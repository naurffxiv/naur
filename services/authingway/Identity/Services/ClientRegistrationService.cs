// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Buffers.Text;
using System.Security.Cryptography;
using Naur.Authingway.Identity.Models;
using OpenIddict.Abstractions;

namespace Naur.Authingway.Identity.Services;

/// <summary>
/// Creates and manages OpenIddict clients exclusively through <see cref="IOpenIddictApplicationManager"/>.
/// </summary>
/// <param name="manager">The OpenIddict application manager.</param>
public sealed class ClientRegistrationService(IOpenIddictApplicationManager manager)
{
    /// <summary>
    /// The number of clients returned per page.
    /// </summary>
    public const int PageSize = 10;

    /// <summary>
    /// Registers a client with a generated identifier, and a generated secret when the client is confidential.
    /// </summary>
    /// <param name="settings">The client settings.</param>
    /// <param name="cancellationToken">The cancellation token.</param>
    /// <returns>The generated client identifier and the plain-text secret, which is not retrievable later.</returns>
    public async Task<ClientCredentials> CreateAsync(ClientSettings settings, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(settings);

        var credentials = new ClientCredentials(
            Guid.NewGuid().ToString("N"),
            settings.IsPublic ? null : GenerateSecret());

        var descriptor = new OpenIddictApplicationDescriptor
        {
            ApplicationType = OpenIddictConstants.ApplicationTypes.Web,
            ClientId = credentials.ClientId,
            ClientSecret = credentials.ClientSecret
        };

        settings.ApplyTo(descriptor);

        await manager.CreateAsync(descriptor, cancellationToken);

        return credentials;
    }

    /// <summary>
    /// Finds a client by its client identifier.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <param name="cancellationToken">The cancellation token.</param>
    /// <returns>The client details, or null when no client exists.</returns>
    public async Task<ClientDetails?> FindAsync(string clientId, CancellationToken cancellationToken = default)
    {
        var application = await manager.FindByClientIdAsync(clientId, cancellationToken);

        if (application is null)
        {
            return null;
        }

        var descriptor = new OpenIddictApplicationDescriptor();
        await manager.PopulateAsync(descriptor, application, cancellationToken);

        return new ClientDetails(clientId, descriptor.ClientType, ClientSettings.FromDescriptor(descriptor));
    }

    /// <summary>
    /// Lists a page of clients.
    /// </summary>
    /// <param name="page">The one-based page number.</param>
    /// <param name="cancellationToken">The cancellation token.</param>
    /// <returns>The requested page and the total number of clients.</returns>
    public async Task<ClientPage> ListAsync(int page, CancellationToken cancellationToken = default)
    {
        ArgumentOutOfRangeException.ThrowIfLessThan(page, 1);

        var total = await manager.CountAsync(cancellationToken);
        var clients = new List<ClientSummary>();

        await foreach (var application in manager.ListAsync(PageSize, (page - 1) * PageSize, cancellationToken))
        {
            clients.Add(new ClientSummary(
                await manager.GetClientIdAsync(application, cancellationToken) ?? string.Empty,
                await manager.GetDisplayNameAsync(application, cancellationToken)));
        }

        return new ClientPage(page, (int)Math.Max(1, (total + PageSize - 1) / PageSize), total, clients);
    }

    /// <summary>
    /// Replaces the managed settings of a client, preserving any unmanaged values. The secret is kept
    /// unless the client becomes public, which removes it, or confidential, which generates a new one.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <param name="settings">The new settings.</param>
    /// <param name="cancellationToken">The cancellation token.</param>
    /// <returns>The client credentials with any newly generated secret, or null when the client does not exist.</returns>
    public async Task<ClientCredentials?> UpdateAsync(string clientId, ClientSettings settings, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(settings);

        var application = await manager.FindByClientIdAsync(clientId, cancellationToken);

        if (application is null)
        {
            return null;
        }

        var descriptor = new OpenIddictApplicationDescriptor();
        await manager.PopulateAsync(descriptor, application, cancellationToken);

        string? generatedSecret = null;

        if (settings.IsPublic)
        {
            descriptor.ClientSecret = null;
        }
        else if (string.IsNullOrEmpty(descriptor.ClientSecret))
        {
            generatedSecret = GenerateSecret();
            descriptor.ClientSecret = generatedSecret;
        }

        settings.ApplyTo(descriptor);

        await manager.UpdateAsync(application, descriptor, cancellationToken);

        return new ClientCredentials(clientId, generatedSecret);
    }

    /// <summary>
    /// Deletes a client.
    /// </summary>
    /// <param name="clientId">The client identifier.</param>
    /// <param name="cancellationToken">The cancellation token.</param>
    /// <returns>True when the client was deleted; false when it does not exist.</returns>
    public async Task<bool> DeleteAsync(string clientId, CancellationToken cancellationToken = default)
    {
        var application = await manager.FindByClientIdAsync(clientId, cancellationToken);

        if (application is null)
        {
            return false;
        }

        await manager.DeleteAsync(application, cancellationToken);

        return true;
    }

    private static string GenerateSecret()
    {
        return Base64Url.EncodeToString(RandomNumberGenerator.GetBytes(32));
    }
}
