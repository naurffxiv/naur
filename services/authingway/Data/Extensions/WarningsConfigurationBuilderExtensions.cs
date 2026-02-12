// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore.Diagnostics;
using System.Reflection;

namespace Naur.Authingway.Data.Extensions;

/// <summary>
/// Provides extension methods for configuring warning behavior in Entity Framework Core, relational, and Npgsql providers.
/// </summary>
public static class WarningsConfigurationBuilderExtensions
{
    /// <summary>
    /// Configures the builder to treat all EF Core, relational, and Npgsql warning events as errors.
    /// </summary>
    /// <remarks>This method updates the warning configuration so that any event classified as a warning by EF
    /// Core, relational, or Npgsql providers will cause an exception to be thrown instead of a warning being logged.
    /// Use this to enforce strict error handling in your application's data access layer.</remarks>
    /// <param name="builder">The builder used to configure warning behavior. Cannot be null.</param>
    public static void TreatWarningsAsErrors(this WarningsConfigurationBuilder builder)
    {
        var eventIdsToThrow = new[] { typeof(CoreEventId), typeof(RelationalEventId), typeof(NpgsqlEfEventId) }
            .SelectMany(t => t.GetFields(BindingFlags.Public | BindingFlags.Static))
            .Where(f => f.FieldType == typeof(EventId))
            .Where(f => f.Name.EndsWith("Warning"))
            .Select(f => (EventId)f.GetValue(null)!)
            .Distinct()
            .ToArray();

        builder.Throw(eventIdsToThrow);
    }
}
