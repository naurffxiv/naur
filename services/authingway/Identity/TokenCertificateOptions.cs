// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;

namespace Naur.Authingway.Identity;

/// <summary>
/// Holds the base64-encoded PKCS#12 certificates OpenIddict uses to sign and encrypt tokens.
/// </summary>
public sealed class TokenCertificateOptions
{
    /// <summary>
    /// The configuration section these options bind from.
    /// </summary>
    public const string SectionName = "OpenIddict";

    /// <summary>
    /// Gets or sets the base64-encoded PKCS#12 signing certificate.
    /// </summary>
    /// <value>The encoded certificate, or <see langword="null"/> or blank when none is configured.</value>
    public string? SigningCertificate { get; set; }

    /// <summary>
    /// Gets or sets the password protecting <see cref="SigningCertificate"/>.
    /// </summary>
    /// <value>The password, or <see langword="null"/> when the certificate is not password protected.</value>
    public string? SigningCertificatePassword { get; set; }

    /// <summary>
    /// Gets or sets the base64-encoded PKCS#12 encryption certificate.
    /// </summary>
    /// <value>The encoded certificate, or <see langword="null"/> or blank when none is configured.</value>
    public string? EncryptionCertificate { get; set; }

    /// <summary>
    /// Gets or sets the password protecting <see cref="EncryptionCertificate"/>.
    /// </summary>
    /// <value>The password, or <see langword="null"/> when the certificate is not password protected.</value>
    public string? EncryptionCertificatePassword { get; set; }

    /// <summary>
    /// Gets a value indicating whether neither certificate is configured.
    /// </summary>
    /// <value><see langword="true"/> when both certificates are absent or blank; otherwise, <see langword="false"/>.</value>
    public bool IsEmpty => string.IsNullOrWhiteSpace(SigningCertificate) && string.IsNullOrWhiteSpace(EncryptionCertificate);

    /// <summary>
    /// Loads the configured signing certificate.
    /// </summary>
    /// <returns>The signing certificate, including its private key.</returns>
    /// <exception cref="InvalidOperationException">The signing certificate is not configured.</exception>
    /// <exception cref="FormatException">The configured value is not valid base64.</exception>
    /// <exception cref="CryptographicException">The certificate cannot be decoded with the configured password.</exception>
    public X509Certificate2 LoadSigningCertificate()
    {
        return Load(SigningCertificate, SigningCertificatePassword, nameof(SigningCertificate));
    }

    /// <summary>
    /// Loads the configured encryption certificate.
    /// </summary>
    /// <returns>The encryption certificate, including its private key.</returns>
    /// <exception cref="InvalidOperationException">The encryption certificate is not configured.</exception>
    /// <exception cref="FormatException">The configured value is not valid base64.</exception>
    /// <exception cref="CryptographicException">The certificate cannot be decoded with the configured password.</exception>
    public X509Certificate2 LoadEncryptionCertificate()
    {
        return Load(EncryptionCertificate, EncryptionCertificatePassword, nameof(EncryptionCertificate));
    }

    private static X509Certificate2 Load(string? encoded, string? password, string name)
    {
        if (string.IsNullOrWhiteSpace(encoded))
        {
            throw new InvalidOperationException($"{SectionName}:{name} is not configured.");
        }

        return X509CertificateLoader.LoadPkcs12(
            Convert.FromBase64String(encoded.Trim()),
            password?.Trim(),
            X509KeyStorageFlags.EphemeralKeySet);
    }
}
