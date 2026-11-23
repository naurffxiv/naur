// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Identity.Models;

/// <summary>
/// A client listed on a page.
/// </summary>
/// <param name="ClientId">The client identifier.</param>
/// <param name="DisplayName">The client display name.</param>
public sealed record ClientSummary(string ClientId, string? DisplayName);
