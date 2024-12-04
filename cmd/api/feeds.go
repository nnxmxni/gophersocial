package main

import (
	"github.com/nnxmxni/gophersocial/types"
	"github.com/nnxmxni/gophersocial/utils"
	"net/http"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	feeds, err := app.store.Posts.GetUserFeed(r.Context(), int64(2))
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	_ = utils.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "User feed retrieved successfully",
		Data:    feeds,
	})
	return
}
