package auth

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var (
	upperRegex   = regexp.MustCompile(`[A-Z]`)
	specialRegex = regexp.MustCompile(`[!@#$%^&*()\-_=+\[\]{};':"\\|,.<>\/?]`)
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,strong_password"`
	Name     string `json:"name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,strong_password"`
}
type GetNewTokensRequest struct {
	Token string `json:"token" validate:"required"`
}

type Tokens struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return upperRegex.MatchString(password) && specialRegex.MatchString(password)
	})
	return v
}
