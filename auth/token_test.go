package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenEndpoint_InvalidGrantType(t *testing.T) {
	auth := NewAuth(authConfig())
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader("grant_type=invalid&client_id=CLIENT_ID&client_secret=client-secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	http.HandlerFunc(auth.Token).ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "unsupported grant type")
}

func TestTokenEndpoint_InvalidClient(t *testing.T) {
	auth := NewAuth(authConfig())
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", "WRONG_ID")
	form.Set("client_secret", "WRONG_SECRET")
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	http.HandlerFunc(auth.Token).ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_client")
}

func TestTokenEndpoint_MissingCode(t *testing.T) {
	auth := NewAuth(authConfig())
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", "CLIENT_ID")
	form.Set("client_secret", "client-secret")
	form.Set("code", "SOME_CODE")
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	http.HandlerFunc(auth.Token).ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "server side error: #3301")
}

func TestTokenEndpoint_HappyFlow_ValidJSON(t *testing.T) {
	auth := NewAuth(authConfig())

	// Generate a real authorization code so oauth2 manager can resolve it from the token store.
	redirectURI := "http://localhost/callback"
	authorizeReq := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	authorizationCode, err := auth.generateCode(authorizeReq.Context(), "code", "user-id", redirectURI, "", authorizeReq)
	assert.NoError(t, err)

	// Token handler still checks the in-memory client->code map before hitting oauth2 manager.
	auth.codes["CLIENT_ID"] = authorizationCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", "CLIENT_ID")
	form.Set("client_secret", "client-secret")
	form.Set("code", authorizationCode)
	form.Set("redirect_uri", redirectURI)

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	http.HandlerFunc(auth.Token).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, w.Body.String(), w.Body.String(), "Response is not valid JSON")
}

// Additional tests for refresh_token and success cases can be added with proper mocking of dependencies.
