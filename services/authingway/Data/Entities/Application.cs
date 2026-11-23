// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using OpenIddict.EntityFrameworkCore.Models;

namespace Naur.Authingway.Data.Entities;

/// <summary>
/// Represents an OpenIddict application entity with GUID identifiers.
/// </summary>
public class Application : OpenIddictEntityFrameworkCoreApplication<Guid, Authorization, Token>, IEntityTypeConfiguration<Application>
{
    /// <inheritdoc/>
    public void Configure(EntityTypeBuilder<Application> builder)
    {
        builder.ToTable("Applications");
    }
}
