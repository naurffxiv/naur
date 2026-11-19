// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Identity.Models;

/// <summary>
/// The credentials generated for a new client.
/// </summary>
/// <param name="ClientId">The client identifier.</param>
/// <param name="ClientSecret">The plain-text client secret.</param>
public sealed record ClientCredentials(string ClientId, string ClientSecret);
