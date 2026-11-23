// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

namespace Naur.Authingway.Identity.Models;

/// <summary>
/// The outcome of validating client settings input.
/// </summary>
/// <param name="Settings">The settings, or null when validation failed.</param>
/// <param name="Errors">The validation errors.</param>
public sealed record ClientSettingsResult(ClientSettings? Settings, IReadOnlyList<string> Errors);
