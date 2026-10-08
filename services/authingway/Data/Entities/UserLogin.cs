// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Naur.Authingway.Data.Entities;

/// <summary>
/// Represents a user login record associated with a user identified by a GUID key.
/// </summary>
public class UserLogin : IdentityUserLogin<Guid>, IEntityTypeConfiguration<UserLogin>
{
    /// <inheritdoc/>
    public void Configure(EntityTypeBuilder<UserLogin> builder)
    {
        builder.ToTable("UserLogins");
    }
}
