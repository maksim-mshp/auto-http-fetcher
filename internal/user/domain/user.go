package domain

import "net/mail"

type User struct {
	ID       int
	Name     string
	Email    mail.Address
	Password string
}
