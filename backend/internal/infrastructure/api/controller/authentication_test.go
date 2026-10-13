package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type whoAmIOutput struct {
	Body struct {
		UserId  string `json:"userId"`
		IsAdmin bool   `json:"isAdmin"`
	}
}

func whoAmI(ctx context.Context, _ *struct{}) (*whoAmIOutput, error) {
	out := &whoAmIOutput{}
	out.Body.UserId = UserIdFromContext(ctx)
	out.Body.IsAdmin = IsAdminFromContext(ctx)
	return out, nil
}

// newAuthTestAPI registers a public, an authenticated and an admin route.
func newAuthTestAPI(t *testing.T) humatest.TestAPI {
	t.Helper()
	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")

	_, api := humatest.New(t)
	api.UseMiddleware(NewAuthenticationMiddleware(api))

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "public-op",
		Path:        "/public",
	}, whoAmI)

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "authenticated-op",
		Path:        "/authenticated",
		Security:    bearerSecurity(),
	}, whoAmI)

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "admin-op",
		Path:        "/admin",
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, whoAmI)

	return api
}

func signTestJWT(t *testing.T, userId, role string, expiresIn time.Duration) string {
	t.Helper()
	claims := domain.AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.String("authentication.jwtSecret")))
	require.NoError(t, err)
	return signed
}

func Test_AuthMiddleware_PublicOp_NoTokenNeeded(t *testing.T) {
	api := newAuthTestAPI(t)

	resp := api.Get("/public")

	require.Equal(t, http.StatusOK, resp.Code)
}

func Test_AuthMiddleware_AuthenticatedOp_MissingBearer_Unauthorized(t *testing.T) {
	api := newAuthTestAPI(t)

	resp := api.Get("/authenticated")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_AuthMiddleware_AuthenticatedOp_InvalidBearer_Unauthorized(t *testing.T) {
	api := newAuthTestAPI(t)

	resp := api.Get("/authenticated", "Authorization: Bearer not-a-real-token")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_AuthMiddleware_AuthenticatedOp_ExpiredBearer_Unauthorized(t *testing.T) {
	api := newAuthTestAPI(t)
	token := signTestJWT(t, "user-1", model.RoleUser, -time.Minute)

	resp := api.Get("/authenticated", "Authorization: Bearer "+token)

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_AuthMiddleware_AuthenticatedOp_ValidBearer_OK(t *testing.T) {
	api := newAuthTestAPI(t)
	token := signTestJWT(t, "user-1", model.RoleUser, 5*time.Minute)

	resp := api.Get("/authenticated", "Authorization: Bearer "+token)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "user-1")
}

func Test_AuthMiddleware_AdminOp_NonAdminBearer_Forbidden(t *testing.T) {
	api := newAuthTestAPI(t)
	token := signTestJWT(t, "user-1", model.RoleUser, 5*time.Minute)

	resp := api.Get("/admin", "Authorization: Bearer "+token)

	require.Equal(t, http.StatusForbidden, resp.Code)
}

func Test_AuthMiddleware_AdminOp_AdminBearer_OK(t *testing.T) {
	api := newAuthTestAPI(t)
	token := signTestJWT(t, "admin-1", model.RoleAdmin, 5*time.Minute)

	resp := api.Get("/admin", "Authorization: Bearer "+token)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "\"isAdmin\":true")
}

const testServiceToken = "service-token-of-at-least-32-bytes-0123456789"

// newServiceTokenTestAPI adds a route the service token may call next to the admin one.
func newServiceTokenTestAPI(t *testing.T, enabled bool) humatest.TestAPI {
	t.Helper()
	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")
	config.Set("authentication.serviceToken.enabled", enabled)
	config.Set("authentication.serviceToken.token", testServiceToken)

	_, api := humatest.New(t)
	api.UseMiddleware(NewAuthenticationMiddleware(api))

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "service-op",
		Path:        "/service",
		Security:    bearerSecurity(),
		Metadata:    requireAdminOrServiceScope("limits"),
	}, whoAmI)
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "admin-op",
		Path:        "/admin",
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, whoAmI)
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "authenticated-op",
		Path:        "/authenticated",
		Security:    bearerSecurity(),
	}, whoAmI)
	return api
}

func Test_AuthMiddleware_ServiceToken_AllowedOnItsOperation(t *testing.T) {
	api := newServiceTokenTestAPI(t, true)

	resp := api.Get("/service", "Authorization: Bearer "+testServiceToken)

	require.Equal(t, http.StatusOK, resp.Code)
}

func Test_AuthMiddleware_ServiceToken_RejectedEverywhereElse(t *testing.T) {
	api := newServiceTokenTestAPI(t, true)

	for _, path := range []string{"/admin", "/authenticated"} {
		resp := api.Get(path, "Authorization: Bearer "+testServiceToken)
		require.Equal(t, http.StatusUnauthorized, resp.Code, path)
	}
}

func Test_AuthMiddleware_ServiceToken_RejectedWhenDisabled(t *testing.T) {
	api := newServiceTokenTestAPI(t, false)

	resp := api.Get("/service", "Authorization: Bearer "+testServiceToken)

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_AuthMiddleware_ServiceToken_WrongTokenRejected(t *testing.T) {
	api := newServiceTokenTestAPI(t, true)

	resp := api.Get("/service", "Authorization: Bearer "+testServiceToken+"x")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_AuthMiddleware_ServiceToken_EnabledWithEmptyTokenMatchesNothing(t *testing.T) {
	api := newServiceTokenTestAPI(t, true)
	config.Set("authentication.serviceToken.token", "")

	resp := api.Get("/service", "Authorization: Bearer ")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_AuthMiddleware_ServiceOp_AdminJwtStillWorks(t *testing.T) {
	api := newServiceTokenTestAPI(t, false)
	token := signTestJWT(t, "admin-1", model.RoleAdmin, 5*time.Minute)

	resp := api.Get("/service", "Authorization: Bearer "+token)

	require.Equal(t, http.StatusOK, resp.Code)
}

func Test_AuthMiddleware_ServiceOp_NonAdminJwtForbidden(t *testing.T) {
	api := newServiceTokenTestAPI(t, true)
	token := signTestJWT(t, "user-1", model.RoleUser, 5*time.Minute)

	resp := api.Get("/service", "Authorization: Bearer "+token)

	require.Equal(t, http.StatusForbidden, resp.Code)
}

func newServiceScopeTestAPI(t *testing.T, scopes ...string) humatest.TestAPI {
	t.Helper()
	config.Reset()
	config.Set("authentication.jwtSecret", "test-secret")
	config.Set("authentication.serviceToken.enabled", true)
	config.Set("authentication.serviceToken.token", testServiceToken)
	if scopes != nil {
		config.Set("authentication.serviceToken.scopes", scopes)
	}

	_, api := humatest.New(t)
	api.UseMiddleware(NewAuthenticationMiddleware(api))

	huma.Register(api, huma.Operation{
		Method: http.MethodGet, OperationID: "limits-op", Path: "/limits",
		Security: bearerSecurity(), Metadata: requireAdminOrServiceScope(config.ScopeLimits),
	}, whoAmI)
	huma.Register(api, huma.Operation{
		Method: http.MethodGet, OperationID: "domains-read-op", Path: "/service/read",
		Security: bearerSecurity(), Metadata: requireServiceScope(config.ScopeDomainsRead),
	}, whoAmI)
	huma.Register(api, huma.Operation{
		Method: http.MethodGet, OperationID: "domains-write-op", Path: "/service/write",
		Security: bearerSecurity(), Metadata: requireServiceScope(config.ScopeDomainsWriteStatus),
	}, whoAmI)
	return api
}

func Test_AuthMiddleware_ServiceScopes_DefaultToLimitsOnly(t *testing.T) {
	api := newServiceScopeTestAPI(t)
	auth := "Authorization: Bearer " + testServiceToken

	require.Equal(t, http.StatusOK, api.Get("/limits", auth).Code)
	require.Equal(t, http.StatusForbidden, api.Get("/service/read", auth).Code)
	require.Equal(t, http.StatusForbidden, api.Get("/service/write", auth).Code)
}

func Test_AuthMiddleware_ServiceScopes_EachRouteNeedsItsOwnScope(t *testing.T) {
	auth := "Authorization: Bearer " + testServiceToken

	api := newServiceScopeTestAPI(t, config.ScopeDomainsRead)
	require.Equal(t, http.StatusOK, api.Get("/service/read", auth).Code)
	require.Equal(t, http.StatusForbidden, api.Get("/service/write", auth).Code)
	require.Equal(t, http.StatusForbidden, api.Get("/limits", auth).Code)

	api = newServiceScopeTestAPI(t, config.ScopeDomainsWriteStatus, config.ScopeLimits)
	require.Equal(t, http.StatusOK, api.Get("/service/write", auth).Code)
	require.Equal(t, http.StatusOK, api.Get("/limits", auth).Code)
	require.Equal(t, http.StatusForbidden, api.Get("/service/read", auth).Code)
}

func Test_AuthMiddleware_ServiceOnlyRoutes_RejectEveryUserToken(t *testing.T) {
	api := newServiceScopeTestAPI(t, config.ScopeDomainsRead)

	for _, role := range []string{model.RoleUser, model.RoleAdmin} {
		token := signTestJWT(t, "someone", role, 5*time.Minute)
		resp := api.Get("/service/read", "Authorization: Bearer "+token)
		require.Equal(t, http.StatusUnauthorized, resp.Code, role)
	}
	require.Equal(t, http.StatusUnauthorized, api.Get("/service/read").Code)
}

func Test_AuthMiddleware_ServiceOnlyRoutes_DisabledTokenMatchesNothing(t *testing.T) {
	api := newServiceScopeTestAPI(t, config.ScopeDomainsRead)
	config.Set("authentication.serviceToken.enabled", false)

	resp := api.Get("/service/read", "Authorization: Bearer "+testServiceToken)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}
