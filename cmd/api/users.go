package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/types"
	"net/http"
	"strconv"
)

type userKey string

const userCtxKey userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	if err := app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "User retrieved successfully",
		Data: map[string]interface{}{
			"user": getUserFromContext(r),
		},
	}); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserFromContext(r)

	toBeFollowedUser, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := app.store.Followers.Follow(r.Context(), user.ID, toBeFollowedUser); err != nil {
		switch {
		case errors.Is(err, store.ErrSelfFollow):
			_ = app.WriteError(w, r, http.StatusConflict, err)
			return
		case errors.Is(err, store.ErrDuplicateFollow):
			_ = app.WriteError(w, r, http.StatusConflict, err)
			return
		default:
			_ = app.WriteError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	_ = app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "User followed successfully",
	})
	return
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	toBeUnfollowedUser, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := app.store.Followers.Unfollow(r.Context(), user.ID, toBeUnfollowedUser); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	_ = app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "successful",
	})
	return
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

	token := chi.URLParam(r, "token")

	ctx := r.Context()
	err := app.store.Users.Activate(ctx, token)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			_ = app.WriteError(w, r, http.StatusNotFound, err)
			return
		default:
			_ = app.WriteError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	_ = app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "Your email has been verified",
	})
	return
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			_ = app.WriteError(w, r, http.StatusBadRequest, err)
			return
		}

		ctx := r.Context()
		user, err := app.store.Users.GetUserByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				_ = app.WriteError(w, r, http.StatusNotFound, err)
				return
			default:
				_ = app.WriteError(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtxKey).(*store.User)
	return user
}
