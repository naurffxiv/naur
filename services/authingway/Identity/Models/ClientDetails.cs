// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Identity.Models;

/// <summary>
/// The stored details of a client, excluding its secret.
/// </summary>
/// <param name="ClientId">The client identifier.</param>
/// <param name="ClientType">The client type.</param>
/// <param name="Settings">The managed client settings.</param>
public sealed record ClientDetails(string ClientId, string? ClientType, ClientSettings Settings);
