package main

import (
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/types"
	"github.com/nnxmxni/gophersocial/utils"
	"net/http"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err = utils.Validate.Struct(fq); err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	feeds, err := app.store.Posts.GetUserFeed(r.Context(), int64(2), fq)
	if err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	_ = app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "User feed retrieved successfully",
		Data:    feeds,
	})
	return
}
