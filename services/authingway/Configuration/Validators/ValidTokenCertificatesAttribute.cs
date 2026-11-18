// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.ComponentModel.DataAnnotations;
using Naur.Authingway.Configuration.Identity;

namespace Naur.Authingway.Configuration.Validators;

/// <summary>
/// Validates that configured OpenIddict certificates are paired and loadable PKCS#12 certificates with private keys.
/// </summary>
[AttributeUsage(AttributeTargets.Class)]
public sealed class ValidTokenCertificatesAttribute : ValidationAttribute
{
    /// <summary>
    /// Initializes a new instance of the <see cref="ValidTokenCertificatesAttribute"/> class.
    /// </summary>
    public ValidTokenCertificatesAttribute()
        : base("OpenIddict signing and encryption certificates must both be valid PKCS#12 certificates with private keys, or both be empty.")
    {
    }

    /// <inheritdoc/>
    public override bool IsValid(object? value)
    {
        return value is IdentityConfig config && config.HasValidCertificates();
    }
}
