// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using Naur.Authingway.Configuration.Discord;
using Naur.Authingway.Configuration.Validators;

namespace Naur.Authingway.Tests.Configuration.Validators;

/// <summary>
/// Tests validation of Discord role dictionaries.
/// </summary>
public class ValidRolesAttributeTests
{
    /// <summary>
    /// Verifies that empty dictionaries and known roles with positive identifiers are accepted.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task AcceptsValidRoles()
    {
        var attribute = new ValidRolesAttribute();

        await Assert.That(attribute.IsValid(new Dictionary<DiscordRole, ulong>())).IsTrue();
        await Assert.That(attribute.IsValid(new Dictionary<DiscordRole, ulong>
        {
            [DiscordRole.Admin] = 1,
            [DiscordRole.Developer] = ulong.MaxValue
        })).IsTrue();
    }

    /// <summary>
    /// Verifies that undefined roles and zero identifiers are rejected even alongside valid entries.
    /// </summary>
    /// <param name="role">The role to validate.</param>
    /// <param name="identifier">The identifier to validate.</param>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    [Arguments((DiscordRole)99, 10UL)]
    [Arguments(DiscordRole.Developer, 0UL)]
    public async Task RejectsInvalidRoles(DiscordRole role, ulong identifier)
    {
        var roles = new Dictionary<DiscordRole, ulong>
        {
            [DiscordRole.Admin] = 1,
            [role] = identifier
        };

        await Assert.That(new ValidRolesAttribute().IsValid(roles)).IsFalse();
    }

    /// <summary>
    /// Verifies that null and incorrectly typed values are rejected.
    /// </summary>
    /// <returns>A task representing the test operation.</returns>
    [Test]
    public async Task RejectsMissingOrIncorrectlyTypedValues()
    {
        var attribute = new ValidRolesAttribute();

        await Assert.That(attribute.IsValid(null)).IsFalse();
        await Assert.That(attribute.IsValid("Admin")).IsFalse();
    }
}
