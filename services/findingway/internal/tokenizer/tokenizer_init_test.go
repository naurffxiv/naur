package tokenizer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeCerts writes a self-signed client cert, key and CA to dir under the names Init reads.
func writeCerts(t *testing.T, dir string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "naur"}, NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	keyDer, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "naur.crt"), certPEM, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "naur.key"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDer}), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hyddwn-ca.crt"), certPEM, 0o600))
}

func withCredentials(t *testing.T) {
	t.Setenv("REDIS_USER", "user")
	t.Setenv("REDIS_PASSWORD", "secret")
}

func TestInitRequiresRedisPassword(t *testing.T) {
	t.Setenv("REDIS_USER", "user")
	t.Setenv("REDIS_PASSWORD", "")
	require.NoError(t, os.Unsetenv("REDIS_PASSWORD"))

	assert.EqualError(t, (&Tokenizer{}).Init(), "REDIS_PASSWORD is not set")
}

func TestInitRequiresCertificates(t *testing.T) {
	withCredentials(t)
	dir := t.TempDir()
	t.Chdir(dir)

	assert.ErrorContains(t, (&Tokenizer{}).Init(), "error loading certificates")

	writeCerts(t, dir)
	require.NoError(t, os.Remove("hyddwn-ca.crt"))
	assert.ErrorContains(t, (&Tokenizer{}).Init(), "error loading CA certificate")

	require.NoError(t, os.WriteFile("hyddwn-ca.crt", []byte("not a cert"), 0o600))
	assert.EqualError(t, (&Tokenizer{}).Init(), "failed to append CA certificate")
}

func TestInitConfiguresTLSRedisClient(t *testing.T) {
	withCredentials(t)
	dir := t.TempDir()
	writeCerts(t, dir)
	t.Chdir(dir)
	tok := &Tokenizer{}

	require.NoError(t, tok.Init())

	opts := tok.rdb.Options()
	assert.Equal(t, "redis.hyddwn.net:6380", opts.Addr)
	assert.Equal(t, "user", opts.Username)
	assert.Equal(t, "secret", opts.Password)
	assert.True(t, opts.ContextTimeoutEnabled)
	require.NotNil(t, opts.TLSConfig)
	assert.Len(t, opts.TLSConfig.Certificates, 1)
}
