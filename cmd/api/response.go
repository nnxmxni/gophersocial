package main

import (
	"encoding/json"
	"github.com/nnxmxni/gophersocial/types"
	"net/http"
)

func (app *application) WriteJSON(w http.ResponseWriter, r *http.Request, status int, response types.APIResponseBody) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	app.logger.Infow(response.Message, "path", r.URL.Path, "method", r.Method)

	return json.NewEncoder(w).Encode(response)
}

func (app *application) WriteError(w http.ResponseWriter, r *http.Request, status int, err error) error {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := types.APIResponseBody{
		Status:  false,
		Message: err.Error(),
	}

	app.logger.Errorw(err.Error(), "path", r.URL.Path, "method", r.Method)

	return json.NewEncoder(w).Encode(response)
}
