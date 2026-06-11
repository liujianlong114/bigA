package api

import (
	"net/http"
	"strings"
)

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			if claims, err := h.Auth.parseToken(tokenStr); err == nil && claims.AccountID > 0 {
				ctx := WithAuth(r.Context(), claims.AccountID, claims.UserID, claims.Username)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		if h.Auth.Relax && h.DefaultAccountID > 0 {
			ctx := WithAuth(r.Context(), h.DefaultAccountID, 0, "default")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		writeErr(w, http.StatusUnauthorized, "login required")
	})
}

func (h *Handlers) accountID(r *http.Request) int64 {
	if id, ok := AccountIDFromContext(r.Context()); ok {
		return id
	}
	return h.DefaultAccountID
}
