// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;
using Naur.Authingway.Configuration.Validators;

namespace Naur.Authingway.Configuration.Identity;

/// <summary>
/// The certificate configuration used by OpenIddict.
/// </summary>
[ValidTokenCertificates]
public sealed class IdentityConfig
{
    /// <summary>
    /// Gets or sets the base64-encoded PKCS#12 signing certificate.
    /// </summary>
    public string? SigningCertificate { get; set; }

    /// <summary>
    /// Gets or sets the password protecting <see cref="SigningCertificate"/>.
    /// </summary>
    public string? SigningCertificatePassword { get; set; }

    /// <summary>
    /// Gets or sets the base64-encoded PKCS#12 encryption certificate.
    /// </summary>
    public string? EncryptionCertificate { get; set; }

    /// <summary>
    /// Gets or sets the password protecting <see cref="EncryptionCertificate"/>.
    /// </summary>
    public string? EncryptionCertificatePassword { get; set; }

    /// <summary>
    /// Gets a value indicating whether neither certificate is configured.
    /// </summary>
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

    internal bool HasValidCertificates()
    {
        if (IsEmpty)
        {
            return true;
        }

        if (string.IsNullOrWhiteSpace(SigningCertificate) || string.IsNullOrWhiteSpace(EncryptionCertificate))
        {
            return false;
        }

        try
        {
            using var signingCertificate = LoadSigningCertificate();
            using var encryptionCertificate = LoadEncryptionCertificate();

            return signingCertificate.HasPrivateKey && encryptionCertificate.HasPrivateKey;
        }
        catch (FormatException)
        {
            return false;
        }
        catch (CryptographicException)
        {
            return false;
        }
    }

    private static X509Certificate2 Load(string? encoded, string? password, string name)
    {
        if (string.IsNullOrWhiteSpace(encoded))
        {
            throw new InvalidOperationException($"Identity:{name} is not configured.");
        }

        return X509CertificateLoader.LoadPkcs12(
            Convert.FromBase64String(encoded.Trim()),
            password?.Trim(),
            X509KeyStorageFlags.EphemeralKeySet);
    }
}
