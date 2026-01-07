package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/VJ-2303/jobhub/internal/data"
	"github.com/VJ-2303/jobhub/internal/mailer"
	"github.com/VJ-2303/jobhub/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &data.User{
		Email:      input.Email,
		Role:       input.Role,
		IsVerified: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	token, err := app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	EmailVerificationData := mailer.VerificationEmailData{
		Email:     user.Email,
		VerifyURL: fmt.Sprintf("http://localhost:4000/user/verify?token=%s", token),
	}

	err = app.mailer.SendVerificationEmail(user.Email, EmailVerificationData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
