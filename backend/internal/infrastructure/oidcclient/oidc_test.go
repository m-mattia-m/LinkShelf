package oidcclient

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/infrastructure/repository"
	"backend/internal/infrastructure/repository/mocks"

	"github.com/coreos/go-oidc/v3/oidc/oidctest"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var assertErr = errors.New("boom")

const (
	fakeClientId = "test-client-id"
	fakeKeyId    = "test-key-1"
)

// fakeProvider is a minimal OIDC provider for tests: discovery + JWKS are
// served exactly to spec (borrowing go-oidc's own oidctest helper for JWKS
// and ID-token signing, since that part is standardized), while /token and
// /userinfo are hand-rolled here since their content is what this package's
// own logic actually has to handle correctly.
type fakeProvider struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey

	idTokenSubject      string
	idTokenExtraClaims  map[string]any
	omitIdToken         bool
	tokenExchangeFails  bool
	userInfoSubject     string // defaults to idTokenSubject when empty
	userInfoClaims      map[string]any
	userInfoUnavailable bool
}

func newFakeProvider(t *testing.T) *fakeProvider {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	fp := &fakeProvider{
		privateKey:     priv,
		idTokenSubject: "user-subject-1",
		userInfoClaims: map[string]any{
			"email":          "user@example.com",
			"email_verified": true,
			"given_name":     "Ada",
			"family_name":    "Lovelace",
		},
	}

	oidcKeys := &oidctest.Server{
		PublicKeys: []oidctest.PublicKey{
			{PublicKey: &priv.PublicKey, KeyID: fakeKeyId, Algorithm: "RS256"},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                fp.server.URL,
			"authorization_endpoint":                fp.server.URL + "/authorize",
			"token_endpoint":                        fp.server.URL + "/token",
			"jwks_uri":                              fp.server.URL + "/keys",
			"userinfo_endpoint":                     fp.server.URL + "/userinfo",
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
		})
	})
	mux.HandleFunc("/keys", oidcKeys.ServeHTTP)
	mux.HandleFunc("/token", fp.serveToken)
	mux.HandleFunc("/userinfo", fp.serveUserInfo)

	fp.server = httptest.NewServer(mux)
	oidcKeys.SetIssuer(fp.server.URL)
	t.Cleanup(fp.server.Close)

	return fp
}

func (fp *fakeProvider) serveToken(w http.ResponseWriter, r *http.Request) {
	if fp.tokenExchangeFails {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		return
	}

	resp := map[string]any{
		"access_token": "fake-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
	}

	if !fp.omitIdToken {
		claims := map[string]any{
			"iss": fp.server.URL,
			"aud": fakeClientId,
			"sub": fp.idTokenSubject,
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}
		for k, v := range fp.idTokenExtraClaims {
			claims[k] = v
		}
		claimsJSON, err := json.Marshal(claims)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp["id_token"] = oidctest.SignIDToken(fp.privateKey, fakeKeyId, "RS256", string(claimsJSON))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (fp *fakeProvider) serveUserInfo(w http.ResponseWriter, r *http.Request) {
	if fp.userInfoUnavailable {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	subject := fp.userInfoSubject
	if subject == "" {
		subject = fp.idTokenSubject
	}

	body := map[string]any{"sub": subject}
	for k, v := range fp.userInfoClaims {
		body[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func setOidcTestConfig(t *testing.T, issuer string) {
	t.Helper()
	config.Reset()
	config.Set("authentication.oidc.issuer", issuer)
	config.Set("authentication.oidc.clientId", fakeClientId)
	config.Set("authentication.oidc.clientSecret", "test-client-secret")
	config.Set("authentication.oidc.redirectUrl", "https://app.example.com/auth/callback")
}

func Test_New_DiscoverySuccess(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)

	client, err := New(context.Background(), stateRepo)

	require.NoError(t, err)
	require.NotNil(t, client)
}

func Test_New_DiscoveryFailure_BadIssuer(t *testing.T) {
	setOidcTestConfig(t, "http://127.0.0.1:1") // nothing listening there

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)

	_, err := New(context.Background(), stateRepo)

	require.Error(t, err)
	require.ErrorContains(t, err, "oidc discovery")
}

func Test_AuthorizationURL_Success(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	authURL, state, err := client.AuthorizationURL()

	require.NoError(t, err)
	require.NotEmpty(t, state)
	require.Contains(t, authURL, fp.server.URL+"/authorize")
	require.Contains(t, authURL, "state="+state)
	require.Contains(t, authURL, "code_challenge=")
	require.Contains(t, authURL, "code_challenge_method=S256")
}

func Test_AuthorizationURL_StateRepoFailure(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(assertErr)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, _, err = client.AuthorizationURL()
	require.ErrorIs(t, err, assertErr)
}

func Test_Exchange_Success(t *testing.T) {
	fp := newFakeProvider(t)
	fp.idTokenSubject = "user-subject-42"
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("the-state").
		Return(&repository.OidcState{
			State:        "the-state",
			CodeVerifier: "a-code-verifier-that-is-long-enough-for-pkce",
			ExpiresAt:    time.Now().Add(time.Minute),
		}, nil)
	stateRepo.EXPECT().
		DeleteByState("the-state").
		Return(nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	identity, err := client.Exchange(context.Background(), "auth-code", "the-state")

	require.NoError(t, err)
	require.NotNil(t, identity)
	require.Equal(t, "user-subject-42", identity.Subject)
	require.Equal(t, "user@example.com", identity.Email)
	require.True(t, identity.EmailVerified)
	require.Equal(t, "Ada", identity.FirstName)
	require.Equal(t, "Lovelace", identity.LastName)
}

func Test_Exchange_UnknownState(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().GetByState("unknown-state").Return(nil, nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "unknown-state")
	require.Error(t, err)
	require.ErrorContains(t, err, "unknown or expired state")
}

func Test_Exchange_ExpiredState(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("expired-state").
		Return(&repository.OidcState{
			State:        "expired-state",
			CodeVerifier: "a-code-verifier-that-is-long-enough-for-pkce",
			ExpiresAt:    time.Now().Add(-time.Minute),
		}, nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "expired-state")
	require.Error(t, err)
	require.ErrorContains(t, err, "unknown or expired state")
}

func Test_Exchange_StateRepoGetFailure(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().GetByState("the-state").Return(nil, assertErr)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "the-state")
	require.ErrorIs(t, err, assertErr)
}

func Test_Exchange_DeleteByStateFailure(t *testing.T) {
	fp := newFakeProvider(t)
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("the-state").
		Return(&repository.OidcState{State: "the-state", CodeVerifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}, nil)
	stateRepo.EXPECT().DeleteByState("the-state").Return(assertErr)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "the-state")
	require.ErrorIs(t, err, assertErr)
}

func Test_Exchange_TokenExchangeFails(t *testing.T) {
	fp := newFakeProvider(t)
	fp.tokenExchangeFails = true
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("the-state").
		Return(&repository.OidcState{State: "the-state", CodeVerifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}, nil)
	stateRepo.EXPECT().DeleteByState("the-state").Return(nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "bad-code", "the-state")
	require.Error(t, err)
	require.ErrorContains(t, err, "token exchange failed")
}

func Test_Exchange_MissingIdToken(t *testing.T) {
	fp := newFakeProvider(t)
	fp.omitIdToken = true
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("the-state").
		Return(&repository.OidcState{State: "the-state", CodeVerifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}, nil)
	stateRepo.EXPECT().DeleteByState("the-state").Return(nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "the-state")
	require.Error(t, err)
	require.ErrorContains(t, err, "did not return an id_token")
}

func Test_Exchange_UserInfoUnavailable(t *testing.T) {
	fp := newFakeProvider(t)
	fp.userInfoUnavailable = true
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("the-state").
		Return(&repository.OidcState{State: "the-state", CodeVerifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}, nil)
	stateRepo.EXPECT().DeleteByState("the-state").Return(nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "the-state")
	require.Error(t, err)
	require.ErrorContains(t, err, "fetching userinfo failed")
}

func Test_Exchange_UserInfoSubjectMismatch_Rejected(t *testing.T) {
	fp := newFakeProvider(t)
	fp.idTokenSubject = "id-token-subject"
	fp.userInfoSubject = "a-different-subject"
	setOidcTestConfig(t, fp.server.URL)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateRepo := mocks.NewMockOidcStateRepository(ctrl)
	stateRepo.EXPECT().
		GetByState("the-state").
		Return(&repository.OidcState{State: "the-state", CodeVerifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}, nil)
	stateRepo.EXPECT().DeleteByState("the-state").Return(nil)

	client, err := New(context.Background(), stateRepo)
	require.NoError(t, err)

	_, err = client.Exchange(context.Background(), "auth-code", "the-state")
	require.Error(t, err)
	require.ErrorContains(t, err, "does not match id_token subject")
}
