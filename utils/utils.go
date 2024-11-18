package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/nnxmxni/gophersocial/types"
	"io"
	"net/http"
)

var Validate = validator.New()

func ParseJSON(w http.ResponseWriter, r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("the request body is empty")
	}

	maxByte := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("the request body is empty")
		}

		return fmt.Errorf("the request body is empty")
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, response types.APIResponseBody) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(response)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, err error) {
	response := types.APIResponseBody{
		Status:  false,
		Message: err.Error(),
	}
	_ = WriteJSON(w, r, status, response)
}
