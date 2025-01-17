package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/types"
	"github.com/nnxmxni/gophersocial/utils"
	"net/http"
	"strconv"
)

type postKey string

const postCtx postKey = "post"

type createPostPayload struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags"`
}

type updatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload createPostPayload

	if err := utils.ParseJSON(w, r, &payload); err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(&payload); err != nil {
		var validationErrors validator.ValidationErrors
		var errorMessages []string
		message := ""
		if errors.As(err, &validationErrors) {
			for _, value := range validationErrors {
				switch value.Tag() {
				case "required":
					message = fmt.Sprintf("%s is %s", value.Field(), value.Tag())
					errorMessages = append(errorMessages, message)
				default:
					message = fmt.Sprintf("The %s is invalid", value.Field())
					errorMessages = append(errorMessages, message)
				}
			}
		}

		_ = app.WriteError(w, r, http.StatusUnprocessableEntity, errors.New(errorMessages[0]))
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  getUserFromContext(r).ID,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := app.WriteJSON(w, r, http.StatusCreated, types.APIResponseBody{
		Status:  true,
		Message: "Post created successfully",
		Data:    post,
	}); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostFromCtx(r)

	comments, err := app.store.Comment.GetByPostID(r.Context(), post.ID)
	if err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	post.Comments = comments

	if err := app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "Post retrieved successfully",
		Data:    post,
	}); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostFromCtx(r)

	var payload updatePostPayload

	if err := utils.ParseJSON(w, r, &payload); err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(&payload); err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "Post updated successfully",
		Data:    post,
	}); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostFromCtx(r)

	if err := app.store.Posts.Delete(r.Context(), post.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			_ = app.WriteError(w, r, http.StatusNotFound, err)
			return
		default:
			_ = app.WriteError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "Post deleted successfully",
	}); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		postIDAsStr := chi.URLParam(r, "postID")
		postIDAsInt, err := strconv.ParseInt(postIDAsStr, 10, 64)
		if err != nil {
			_ = app.WriteError(w, r, http.StatusBadRequest, fmt.Errorf("invalid post id - %s", postIDAsStr))
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.GetPostByID(ctx, postIDAsInt)
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

		ctx = context.WithValue(r.Context(), postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
