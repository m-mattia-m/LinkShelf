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

// newAuthTestAPI registers the same middleware Router() wires up, plus three
// operations mirroring the three security shapes actually used in
// router.go: public (no Security), any authenticated caller (bearerSecurity),
// and admin-only (bearerSecurity + requireAdmin).
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
