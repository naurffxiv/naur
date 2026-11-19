// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Identity.Models;

/// <summary>
/// A page of clients.
/// </summary>
/// <param name="Page">The one-based page number.</param>
/// <param name="PageCount">The total number of pages.</param>
/// <param name="Total">The total number of clients.</param>
/// <param name="Clients">The clients on this page.</param>
public sealed record ClientPage(int Page, int PageCount, long Total, IReadOnlyList<ClientSummary> Clients);
