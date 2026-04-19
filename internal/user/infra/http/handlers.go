package http

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/user/domain"
	"auto-http-fetcher/internal/user/service"
	"errors"
	"log/slog"
	"net/http"
)

type UserHandlers struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewUserHandlers(logger *slog.Logger, userService *service.UserService) *UserHandlers {
	return &UserHandlers{userService, logger}
}

func (uh *UserHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var LoginReq UserLoginRequest
	if err := coreHttp.ParseJSONBody(uh.logger, r, &LoginReq); err != nil {
		coreHttp.SendErrorJSON(uh.logger, w, err)
		return
	}

	token, err := uh.userService.Get(r.Context(), &domain.User{
		Email:    LoginReq.Email,
		Password: LoginReq.Password,
	})
	if err != nil {
		var apiErr coreHttp.APIError
		if errors.As(err, &apiErr) {
			coreHttp.SendErrorJSON(uh.logger, w, &apiErr)
			return
		}
		coreHttp.SendErrorJSON(uh.logger, w, &coreHttp.ErrInternal)
		return
	}

	coreHttp.SendJSON(uh.logger, w, &UserLoginResponse{
		Token: token,
	}, http.StatusOK)
}

func (uh *UserHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var RegisterReq UserRegisterRequest
	if err := coreHttp.ParseJSONBody(uh.logger, r, &RegisterReq); err != nil {
		coreHttp.SendJSON(uh.logger, w, err, http.StatusBadRequest)
		return
	}
	user, err := uh.userService.Create(r.Context(), &domain.User{
		Email:    RegisterReq.Email,
		Password: RegisterReq.Password,
		Name:     RegisterReq.Name,
	})
	if err != nil {
		var apiErr coreHttp.APIError
		if errors.As(err, &apiErr) {
			coreHttp.SendErrorJSON(uh.logger, w, &apiErr)
			return
		}
		coreHttp.SendErrorJSON(uh.logger, w, &coreHttp.ErrInternal)
		return
	}

	coreHttp.SendJSON(uh.logger, w, &UserRegisterResponse{
		Email: user.Email,
		Name:  user.Name,
	}, http.StatusCreated)
}
