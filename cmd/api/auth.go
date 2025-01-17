package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nnxmxni/gophersocial/internals/store"
	"github.com/nnxmxni/gophersocial/types"
	"github.com/nnxmxni/gophersocial/utils"
	"net/http"
	"time"
)

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var payload RegisterUserPayload
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

		_ = app.WriteError(w, r, http.StatusBadRequest, errors.New(errorMessages[0]))
		return
	}

	user := &store.User{
		Email: payload.Email,
		Role: store.Roles{
			Name: "user",
		},
	}

	if err := user.Password.Set(payload.Password); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	plainToken := uuid.New().String()

	// hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := app.store.Users.CreateAndInvite(r.Context(), user, hashToken, app.config.mail.OTPExpiration); err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	_ = app.WriteJSON(w, r, http.StatusCreated, types.APIResponseBody{
		Status:  true,
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"token": plainToken,
			"user":  user,
		},
	})
	return
}

func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {

	var payload LoginUserPayload
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
					message = fmt.Sprintf("The %s is invalid", value.Tag())
					errorMessages = append(errorMessages, message)
				}
			}
		}

		_ = app.WriteError(w, r, http.StatusUnprocessableEntity, errors.New(errorMessages[0]))
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			err = errors.New("incorrect email or password")
		default:
			_ = app.WriteError(w, r, http.StatusInternalServerError, err)
			return
		}
		_ = app.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		_ = app.WriteError(w, r, http.StatusBadRequest, errors.New("incorrect email or password"))
		return
	}

	token, err := app.authenticator.GenerateToken(
		jwt.MapClaims{
			"sub": user.ID,
			"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"iss": app.config.auth.token.host,
			"aud": app.config.auth.token.host,
		},
	)
	if err != nil {
		_ = app.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	_ = app.WriteJSON(w, r, http.StatusOK, types.APIResponseBody{
		Status:  true,
		Message: "Welcome back",
		Token:   token,
		Data: map[string]interface{}{
			"user": user,
		},
	})
	return
}
