package http

type UserRegisterRequest struct {
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}
type UserRegisterResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserLoginResponse struct {
	Token string `json:"token"`
}
