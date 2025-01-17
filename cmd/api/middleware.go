package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nnxmxni/gophersocial/internals/store"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) EnsureAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			_ = app.WriteError(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			_ = app.WriteError(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		token := parts[1]
		validatedToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			_ = app.WriteError(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		claims, _ := validatedToken.Claims.(jwt.MapClaims)

		userID, _ := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		ctx := r.Context()

		user, err := app.getUser(ctx, userID)
		if err != nil {
			_ = app.WriteError(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		if err != nil {
			_ = app.WriteError(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) EnsurePostOwnership(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user := getUserFromContext(r)
		post := getPostFromCtx(r)

		if user.ID != post.UserID {
			_ = app.WriteError(w, r, http.StatusForbidden, errors.New("forbidden"))
			return
		}

		allowed, err := app.confirmRolePrecedence(r.Context(), user, role)
		if err != nil {
			_ = app.WriteError(w, r, http.StatusInternalServerError, err)
			return
		}

		if !allowed {
			_ = app.WriteError(w, r, http.StatusForbidden, errors.New("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (app *application) confirmRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {

	if !app.config.redisCfg.enabled {
		return app.store.Users.GetUserByID(ctx, userID)
	}

	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil {
		panic(err)
		return nil, err
	}

	if user == nil {
		user, err = app.store.Users.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := app.cacheStorage.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {

			ip := r.RemoteAddr
			// Strip the port if present
			if host, _, err := net.SplitHostPort(ip); err == nil {
				ip = host
			}

			// Check X-Forwarded-For header
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// Use the first IP in the list
				ips := strings.Split(xff, ",")
				ip = strings.TrimSpace(ips[0])
			}

			if allow, retryAfter := app.rateLimiter.Allow(ip); !allow {
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
				_ = app.WriteError(w, r, http.StatusTooManyRequests, errors.New("too many requests"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
