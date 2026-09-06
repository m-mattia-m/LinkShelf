package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type contextKey string

const (
	userIdContextKey contextKey = "userId"
	roleContextKey   contextKey = "role"

	// roleMetadataKey, set on an operation's Metadata, restricts it to callers
	// with that exact role. Operations without it just require a valid token.
	roleMetadataKey = "requiredRole"
)

// UserIdFromContext returns the authenticated caller's user ID, or "" if the
// request reached an unauthenticated (public) operation.
func UserIdFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIdContextKey).(string)
	return v
}

// IsAdminFromContext reports whether the authenticated caller has the admin role.
func IsAdminFromContext(ctx context.Context) bool {
	role, _ := ctx.Value(roleContextKey).(string)
	return role == model.RoleAdmin
}

// NewAuthenticationMiddleware validates our own JWT for any operation that
// declares a Security requirement. Operations without one (Security == nil,
// the huma default) are public and skip authentication entirely. This is the
// only check performed regardless of authentication.type - LOCAL and OIDC
// logins both end up with the same kind of token.
func NewAuthenticationMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		op := ctx.Operation()
		if op == nil || len(op.Security) == 0 {
			next(ctx)
			return
		}

		bearer := strings.TrimSpace(strings.TrimPrefix(ctx.Header("Authorization"), "Bearer "))
		if bearer == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := domain.ValidateAccessToken(bearer)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		if requiredRole, ok := op.Metadata[roleMetadataKey].(string); ok && requiredRole != "" && claims.Role != requiredRole {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "this endpoint requires the "+requiredRole+" role")
			return
		}

		ctx = huma.WithValue(ctx, userIdContextKey, claims.Subject)
		ctx = huma.WithValue(ctx, roleContextKey, claims.Role)

		next(ctx)
	}
}

// requireAdmin marks an operation as admin-only. Pass into huma.Operation{Metadata: ...}.
func requireAdmin() map[string]any {
	return map[string]any{roleMetadataKey: model.RoleAdmin}
}

// bearerSecurity marks an operation as requiring a valid access token.
func bearerSecurity() []map[string][]string {
	return []map[string][]string{{"bearer": {}}}
}
