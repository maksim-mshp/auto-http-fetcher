package http

type UserRegisterRequest struct {
	Password string `json:"password" example:"strong-password"`
	Name     string `json:"name" example:"Maksim"`
	Email    string `json:"email" example:"user@example.com"`
}
type UserRegisterResponse struct {
	Email string `json:"email" example:"user@example.com"`
	Name  string `json:"name" example:"Maksim"`
}

type UserLoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"strong-password"`
}
type UserLoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOi..."`
}
