package domain

import (
	"fmt"
	"regexp"
)

const (
	RFC5322EmailPattern   = `^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	UsernameMinLength     = 3
	UsernameMaxLength     = 255
	UserPasswordMinLength = 6
)

func ValidateUser(user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	if user.Name == "" {
		return fmt.Errorf("user's name is empty")
	}

	if len(user.Name) < UsernameMinLength {
		return fmt.Errorf("user's name is too short")
	}
	if len(user.Name) > UsernameMaxLength {
		return fmt.Errorf("user's name is too long")
	}

	if match, _ := regexp.MatchString(RFC5322EmailPattern, user.Email); !match {
		return fmt.Errorf("invalid email")
	}

	if user.Password == "" {
		return fmt.Errorf("user's password is empty")
	}
	if len(user.Password) < UserPasswordMinLength {
		return fmt.Errorf("password too short")
	}
	return nil
}
