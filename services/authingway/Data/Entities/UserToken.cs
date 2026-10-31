// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Naur.Authingway.Data.Entities;

/// <summary>
/// Represents an authentication token associated with a user, using a GUID as the user identifier.
/// </summary>
public class UserToken : IdentityUserToken<Guid>, IEntityTypeConfiguration<UserToken>
{
    /// <inheritdoc/>
    public void Configure(EntityTypeBuilder<UserToken> builder)
    {
        builder.ToTable("UserTokens");
    }
}
