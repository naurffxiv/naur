package fflogs

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tokenEndpoint struct {
	status int
	body   string
	auth   string
}

func (e *tokenEndpoint) RoundTrip(r *http.Request) (*http.Response, error) {
	e.auth = r.Header.Get("Authorization")
	return &http.Response{StatusCode: e.status, Body: io.NopCloser(strings.NewReader(e.body)), Request: r}, nil
}

func withTokenEndpoint(t *testing.T, e *tokenEndpoint) {
	t.Helper()
	old := http.DefaultTransport
	http.DefaultTransport = e
	t.Cleanup(func() { http.DefaultTransport = old })
}

func TestTokenUsesClientCredentials(t *testing.T) {
	e := &tokenEndpoint{status: http.StatusOK, body: `{"access_token":"abc","token_type":"Bearer","expires_in":60}`}
	withTokenEndpoint(t, e)

	token, err := Init("id", "secret").Token()

	require.NoError(t, err)
	assert.Equal(t, "abc", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.True(t, token.Valid())
	assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("id:secret")), e.auth)
}

func TestTokenErrors(t *testing.T) {
	tests := map[string]struct {
		status int
		body   string
		want   string
	}{
		"http error":    {http.StatusUnauthorized, ``, "fflogs oauth2 token request returned HTTP 401"},
		"oauth error":   {http.StatusOK, `{"error":"invalid_client","error_description":"bad secret"}`, "token error invalid_client: bad secret"},
		"missing token": {http.StatusOK, `{"token_type":"Bearer"}`, "fflogs oauth2 token response contained no access token"},
		"not json":      {http.StatusOK, `nope`, "could not coerce response to JSON"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			withTokenEndpoint(t, &tokenEndpoint{status: tt.status, body: tt.body})

			_, err := Init("id", "secret").Token()

			assert.ErrorContains(t, err, tt.want)
		})
	}
}
