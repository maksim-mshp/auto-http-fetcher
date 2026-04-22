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

type APIError = coreHttp.APIError

func NewUserHandlers(logger *slog.Logger, userService *service.UserService) *UserHandlers {
	return &UserHandlers{userService, logger}
}

// Login godoc
// @Summary		Войти в аккаунт
// @Description	Проверяет email и пароль, затем возвращает JWT для защищённых методов.
// @Tags		Авторизация
// @Accept		json
// @Produce		json
// @Param		request body UserLoginRequest true "Учётные данные"
// @Success		200 {object} UserLoginResponse
// @Failure		400 {object} APIError
// @Failure		403 {object} APIError
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/auth/login [post]
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

// Register godoc
// @Summary		Зарегистрировать пользователя
// @Description	Создаёт пользователя и возвращает публичные данные созданного аккаунта.
// @Tags		Авторизация
// @Accept		json
// @Produce		json
// @Param		request body UserRegisterRequest true "Данные пользователя"
// @Success		201 {object} UserRegisterResponse
// @Failure		400 {object} APIError
// @Failure		409 {object} APIError
// @Failure		500 {object} APIError
// @Router		/auth/register [post]
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
