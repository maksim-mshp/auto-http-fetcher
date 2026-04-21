package domain

import "fmt"

type NotFoundError struct {
	ID int
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("item with id=%d not found", e.ID)
}

type AlreadyExistsError struct {
	ID int
}

func (e AlreadyExistsError) Error() string {
	return fmt.Sprintf("item with id=%d already exists", e.ID)
}
