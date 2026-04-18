package security

import "golang.org/x/crypto/bcrypt"

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func Verify(hashedPassword, comparePassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(comparePassword))
	return err == nil
}
