package main

import (
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

type createPostPayload struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload createPostPayload

	if err := utils.ParseJSON(w, r, &payload); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err)
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

		utils.WriteError(w, r, http.StatusUnprocessableEntity, errors.New(errorMessages[0]))
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := utils.WriteJSON(w, r, http.StatusCreated, types.APIResponseBody{
		Status:  true,
		Message: "Post created successfully",
		Data:    post,
	}); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	postIDAsStr := chi.URLParam(r, "postID")
	postIDAsInt, err := strconv.ParseInt(postIDAsStr, 10, 64)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, fmt.Errorf("invalid post id - %s", postIDAsStr))
		return
	}

	ctx := r.Context()
	post, err := app.store.Posts.GetPostByID(ctx, postIDAsInt)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			utils.WriteError(w, r, http.StatusNotFound, err)
			return
		default:
			utils.WriteError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := utils.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "Post retrieved successfully",
		Data:    post,
	}); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}
