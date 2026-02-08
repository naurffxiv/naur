// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using FastEndpoints;

namespace Naur.Authingway.Core.Endpoints;

/// <summary>
/// Represents an endpoint that redirects incoming requests to the documentation page.
/// </summary>
public class IndexEndpoint : EndpointWithoutRequest
{
    /// <inheritdoc/>
    public override void Configure()
    {
        Get("/");
        AllowAnonymous();
        Description(b =>
        {
            b.ExcludeFromDescription();
        });
    }

    /// <inheritdoc/>
    public override async Task HandleAsync(CancellationToken ct)
    {
        await Send.RedirectAsync("/docs");
    }
}
