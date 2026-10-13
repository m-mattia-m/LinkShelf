package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type contextKey string

const (
	userIdContextKey  contextKey = "userId"
	roleContextKey    contextKey = "role"
	serviceContextKey contextKey = "serviceToken"

	// roleMetadataKey restricts an operation to a role.
	roleMetadataKey = "requiredRole"
	// serviceScopeMetadataKey lets the service token call an operation, if it holds this scope.
	serviceScopeMetadataKey = "serviceScope"
	// serviceOnlyMetadataKey marks an operation that only the service token may call.
	serviceOnlyMetadataKey = "serviceOnly"
)

// UserIdFromContext returns the caller's user ID, or "" if unauthenticated.
func UserIdFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIdContextKey).(string)
	return v
}

// IsAdminFromContext reports whether the authenticated caller has the admin role.
func IsAdminFromContext(ctx context.Context) bool {
	role, _ := ctx.Value(roleContextKey).(string)
	return role == model.RoleAdmin
}

// IsServiceTokenFromContext reports whether the caller authenticated with the service token.
func IsServiceTokenFromContext(ctx context.Context) bool {
	v, _ := ctx.Value(serviceContextKey).(bool)
	return v
}

// matchesServiceToken compares in constant time, regardless of length.
func matchesServiceToken(bearer string) bool {
	if !config.ServiceTokenEnabled() {
		return false
	}
	token := config.ServiceToken()
	if token == "" {
		return false
	}
	a, b := sha256.Sum256([]byte(bearer)), sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

// NewAuthenticationMiddleware validates the JWT on operations with a Security requirement.
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

		if matchesServiceToken(bearer) {
			scope, _ := op.Metadata[serviceScopeMetadataKey].(string)
			if scope == "" {
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			if !config.ServiceTokenHasScope(scope) {
				_ = huma.WriteErr(api, ctx, http.StatusForbidden, "the service token lacks the scope "+scope)
				return
			}
			ctx = huma.WithValue(ctx, userIdContextKey, "service")
			ctx = huma.WithValue(ctx, roleContextKey, model.RoleAdmin)
			ctx = huma.WithValue(ctx, serviceContextKey, true)
			next(ctx)
			return
		}

		if serviceOnly, _ := op.Metadata[serviceOnlyMetadataKey].(bool); serviceOnly {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
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

// requireAdminOrServiceScope is requireAdmin, also open to the service token with this scope.
func requireAdminOrServiceScope(scope string) map[string]any {
	return map[string]any{roleMetadataKey: model.RoleAdmin, serviceScopeMetadataKey: scope}
}

// requireServiceScope marks an operation that only the service token with this scope may call.
// A user token is rejected, so the service routes can never be reached as a user.
func requireServiceScope(scope string) map[string]any {
	return map[string]any{serviceOnlyMetadataKey: true, serviceScopeMetadataKey: scope}
}

// bearerSecurity marks an operation as requiring a valid access token.
func bearerSecurity() []map[string][]string {
	return []map[string][]string{{"bearer": {}}}
}
