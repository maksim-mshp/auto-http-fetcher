package domain

import (
	coreHttp "auto-http-fetcher/internal/core/http"
)

const (
	UsernameMinLength     = 4
	UsernameMaxLength     = 128
	UserPasswordMinLength = 8
)

func ValidateUser(user *User) error {
	if user == nil {
		return coreHttp.NewValidationError("user", "user is nil")
	}

	if len(user.Name) < UsernameMinLength {
		return coreHttp.NewValidationError("username", "username is too short")
	}
	if len(user.Name) > UsernameMaxLength {
		return coreHttp.NewValidationError("username", "username is too long")
	}

	if len(user.Password) < UserPasswordMinLength {
		return coreHttp.NewValidationError("user password", "password is too short")
	}
	return nil
}
