package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/types"
	"github.com/nnxmxni/gophersocial/utils"
	"net/http"
	"strconv"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	user, err := app.store.Users.GetUserByID(ctx, userID)
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

	if err = utils.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "User retrieved successfully",
		Data:    user,
	}); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}
}
