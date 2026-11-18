// Licensed to the NAUR Contributors under one or more agreements.
// The NAUR Contributors licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.

using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Options;
using Naur.Authingway.Configuration;
using Naur.Authingway.Configuration.Identity;

namespace Naur.Authingway.Tests.Configuration;

/// <summary>
/// Tests Identity certificate configuration and validation through the configured options pipeline.
/// </summary>
public class IdentityConfigTests
{
    private const string Password = "test-password";

    /// <summary>
    /// Verifies that configured certificates load with their private keys.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task LoadsConfiguredCertificatesWithPrivateKeys()
    {
        using var signingSource = CreateCertificate(X509KeyUsageFlags.DigitalSignature);
        using var encryptionSource = CreateCertificate(X509KeyUsageFlags.KeyEncipherment);
        var config = new IdentityConfig
        {
            SigningCertificate = Encode(signingSource, Password),
            SigningCertificatePassword = Password,
            EncryptionCertificate = Encode(encryptionSource, Password),
            EncryptionCertificatePassword = Password
        };

        using var signingCertificate = config.LoadSigningCertificate();
        using var encryptionCertificate = config.LoadEncryptionCertificate();

        await Assert.That(signingCertificate.Thumbprint).IsEqualTo(signingSource.Thumbprint);
        await Assert.That(signingCertificate.HasPrivateKey).IsTrue();
        await Assert.That(encryptionCertificate.Thumbprint).IsEqualTo(encryptionSource.Thumbprint);
        await Assert.That(encryptionCertificate.HasPrivateKey).IsTrue();
        await StartHost(new Dictionary<string, string?>
        {
            ["Identity:SigningCertificate"] = config.SigningCertificate,
            ["Identity:SigningCertificatePassword"] = config.SigningCertificatePassword,
            ["Identity:EncryptionCertificate"] = config.EncryptionCertificate,
            ["Identity:EncryptionCertificatePassword"] = config.EncryptionCertificatePassword
        });
    }

    /// <summary>
    /// Verifies that an empty certificate configuration is accepted.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task AcceptsEmptyCertificateConfiguration()
    {
        await StartHost([]);
    }

    /// <summary>
    /// Verifies that partial certificate configuration is rejected.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RejectsPartialCertificateConfiguration()
    {
        await Assert.That(async () => await StartHost(new Dictionary<string, string?>
        {
            ["Identity:SigningCertificate"] = "value"
        })).Throws<OptionsValidationException>();
    }

    /// <summary>
    /// Verifies that malformed PKCS#12 data and incorrect passwords fail options validation at startup.
    /// </summary>
    /// <returns>A task that represents the asynchronous test operation.</returns>
    [Test]
    public async Task RejectsInvalidCertificatesOnStartup()
    {
        await Assert.That(async () => await StartHost(new Dictionary<string, string?>
        {
            ["Identity:SigningCertificate"] = "not base64!",
            ["Identity:EncryptionCertificate"] = "not base64!"
        })).Throws<OptionsValidationException>();

        using var source = CreateCertificate(X509KeyUsageFlags.DigitalSignature);
        var encoded = Encode(source, Password);
        await Assert.That(async () => await StartHost(new Dictionary<string, string?>
        {
            ["Identity:SigningCertificate"] = encoded,
            ["Identity:SigningCertificatePassword"] = "wrong-password",
            ["Identity:EncryptionCertificate"] = encoded,
            ["Identity:EncryptionCertificatePassword"] = "wrong-password"
        })).Throws<OptionsValidationException>();
    }

    private static async Task StartHost(Dictionary<string, string?> identityConfiguration)
    {
        var settings = new HostApplicationBuilderSettings { DisableDefaults = true };
        var builder = Host.CreateApplicationBuilder(settings);
        identityConfiguration["Discord:ClientId"] = "123456789";
        identityConfiguration["Discord:ClientSecret"] = "01234567890123456789012345678901";
        identityConfiguration["Discord:GuildId"] = "1";
        builder.Configuration.AddInMemoryCollection(identityConfiguration);
        builder.ConfigureConfiguration();

        using var host = builder.Build();
        await host.StartAsync();
        await host.StopAsync();
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
