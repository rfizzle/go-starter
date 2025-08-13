package auth

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// RoleProvider interface for fetching user roles
type RoleProvider interface {
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

// Permission represents a permission check
type Permission struct {
	Resource string
	Action   string
}

// AuthorizationProvider interface for checking permissions
type AuthorizationProvider interface {
	HasPermission(ctx context.Context, userID string, permission Permission) (bool, error)
}

// RequireRole middleware that checks if user has required role
func (a *Auth) RequireRole(roleProvider RoleProvider, requiredRole string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First ensure user is authenticated
			subject, err := a.subjectFromContext(r.Context())
			if err != nil {
				a.logger.Error("rbac: failed to get subject from context", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if subject == "" {
				a.logger.Error("rbac: empty subject")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get user roles
			roles, err := roleProvider.GetUserRoles(r.Context(), subject)
			if err != nil {
				a.logger.Error("rbac: failed to get user roles",
					zap.String("subject", subject), zap.Error(err))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				a.logger.Warn("rbac: user lacks required role",
					zap.String("subject", subject),
					zap.String("required_role", requiredRole),
					zap.Strings("user_roles", roles))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission middleware that checks if user has required permission
func (a *Auth) RequirePermission(authzProvider AuthorizationProvider, permission Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First ensure user is authenticated
			subject, err := a.subjectFromContext(r.Context())
			if err != nil {
				a.logger.Error("authz: failed to get subject from context", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if subject == "" {
				a.logger.Error("authz: empty subject")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check permission
			hasPermission, err := authzProvider.HasPermission(r.Context(), subject, permission)
			if err != nil {
				a.logger.Error("authz: failed to check permission",
					zap.String("subject", subject),
					zap.String("resource", permission.Resource),
					zap.String("action", permission.Action),
					zap.Error(err))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !hasPermission {
				a.logger.Warn("authz: user lacks required permission",
					zap.String("subject", subject),
					zap.String("resource", permission.Resource),
					zap.String("action", permission.Action))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole middleware that checks if user has any of the required roles
func (a *Auth) RequireAnyRole(roleProvider RoleProvider, requiredRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject, err := a.subjectFromContext(r.Context())
			if err != nil {
				a.logger.Error("rbac: failed to get subject from context", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if subject == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			roles, err := roleProvider.GetUserRoles(r.Context(), subject)
			if err != nil {
				a.logger.Error("rbac: failed to get user roles",
					zap.String("subject", subject), zap.Error(err))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, userRole := range roles {
				for _, requiredRole := range requiredRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				a.logger.Warn("rbac: user lacks any required role",
					zap.String("subject", subject),
					zap.Strings("required_roles", requiredRoles),
					zap.Strings("user_roles", roles))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
