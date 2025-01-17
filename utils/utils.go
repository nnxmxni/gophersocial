package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
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
