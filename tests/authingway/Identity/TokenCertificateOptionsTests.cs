// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;
using Naur.Authingway.Identity;

namespace Naur.Authingway.Tests.Identity;

/// <summary>
/// Contains tests that verify token certificates load from their configured base64 PKCS#12 form.
/// </summary>
public class TokenCertificateOptionsTests
{
    private const string Password = "test-password";

    /// <summary>
    /// Verifies that a configured signing certificate loads with its private key.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task LoadsSigningCertificateWithPrivateKey()
    {
        using var source = CreateCertificate(X509KeyUsageFlags.DigitalSignature);
        var options = new TokenCertificateOptions
        {
            SigningCertificate = Encode(source, Password),
            SigningCertificatePassword = Password
        };

        using var loaded = options.LoadSigningCertificate();

        await Assert.That(loaded.Thumbprint).IsEqualTo(source.Thumbprint);
        await Assert.That(loaded.HasPrivateKey).IsTrue();
    }

    /// <summary>
    /// Verifies that a configured encryption certificate loads with its private key.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task LoadsEncryptionCertificateWithPrivateKey()
    {
        using var source = CreateCertificate(X509KeyUsageFlags.KeyEncipherment);
        var options = new TokenCertificateOptions
        {
            EncryptionCertificate = Encode(source, Password),
            EncryptionCertificatePassword = Password
        };

        using var loaded = options.LoadEncryptionCertificate();

        await Assert.That(loaded.Thumbprint).IsEqualTo(source.Thumbprint);
        await Assert.That(loaded.HasPrivateKey).IsTrue();
    }

    /// <summary>
    /// Verifies that loading an unconfigured certificate fails with the missing configuration key.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesUnconfiguredCertificate()
    {
        var options = new TokenCertificateOptions();

        await Assert.That(() => options.LoadSigningCertificate())
            .Throws<InvalidOperationException>()
            .WithMessageContaining("OpenIddict:SigningCertificate");
    }

    /// <summary>
    /// Verifies that a certificate fails to load with the wrong password.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesWrongPassword()
    {
        using var source = CreateCertificate(X509KeyUsageFlags.DigitalSignature);
        var options = new TokenCertificateOptions
        {
            SigningCertificate = Encode(source, Password),
            SigningCertificatePassword = "wrong-password"
        };

        await Assert.That(() => options.LoadSigningCertificate())
            .Throws<CryptographicException>();
    }

    /// <summary>
    /// Verifies that a value that is not base64 fails to load.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RefusesInvalidBase64()
    {
        var options = new TokenCertificateOptions
        {
            SigningCertificate = "not base64!",
            SigningCertificatePassword = Password
        };

        await Assert.That(() => options.LoadSigningCertificate())
            .Throws<FormatException>();
    }

    /// <summary>
    /// Verifies that options with no certificates report as empty and options with one report as configured.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task ReportsWhetherAnyCertificateIsConfigured()
    {
        var empty = new TokenCertificateOptions { SigningCertificate = " ", EncryptionCertificate = "" };
        var partial = new TokenCertificateOptions { SigningCertificate = "value" };

        await Assert.That(empty.IsEmpty).IsTrue();
        await Assert.That(partial.IsEmpty).IsFalse();
    }

    private static X509Certificate2 CreateCertificate(X509KeyUsageFlags usage)
    {
        using var key = RSA.Create(2048);
        var request = new CertificateRequest("CN=Authingway Test", key, HashAlgorithmName.SHA256, RSASignaturePadding.Pkcs1);
        request.CertificateExtensions.Add(new X509KeyUsageExtension(usage, critical: true));

        return request.CreateSelfSigned(DateTimeOffset.UtcNow.AddMinutes(-5), DateTimeOffset.UtcNow.AddYears(1));
    }

    private static string Encode(X509Certificate2 certificate, string password)
    {
        return Convert.ToBase64String(certificate.Export(X509ContentType.Pkcs12, password));
    }
}
