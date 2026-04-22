package main

import (
	"embed"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//go:embed api/openapi.json api/openapi.yml
var openAPIFS embed.FS

func RegisterSwagger(mux *http.ServeMux) {
	mux.HandleFunc("/swagger/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, openAPIFS, "api/openapi.json")
	})
	mux.HandleFunc("/swagger/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, openAPIFS, "api/openapi.yml")
	})
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/openapi.json"),
	))
}

type FetchHeaderValues struct {
	Values []string `json:"values" example:"application/json"`
} // @name FetchHeaderValues

type APIError struct {
	Error   string                 `json:"error" example:"INVALID_BODY"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object"`
} // @name APIError

type AnalyticsResponse struct {
	TotalCalls   int             `json:"TotalCalls" example:"42"`
	SuccessCalls int             `json:"SuccessCalls" example:"37"`
	FailedCalls  int             `json:"FailedCalls" example:"5"`
	StatusStats  map[int]float64 `json:"StatusStats"`
	AvgDuration  int64           `json:"AvgDuration" example:"250000000"`
	MinDuration  int64           `json:"MinDuration" example:"100000000"`
	MaxDuration  int64           `json:"MaxDuration" example:"900000000"`
	AvgAttempts  float64         `json:"AvgAttempts" example:"1.2"`
} // @name AnalyticsResponse

type FetchRequest struct {
	ID          int64                        `json:"id" example:"1"`
	Description string                       `json:"description" example:"Health check"`
	IntervalMs  int64                        `json:"intervalMs" example:"60000"`
	TimeoutMs   int64                        `json:"timeoutMs" example:"5000"`
	URL         string                       `json:"url" example:"https://example.com/health"`
	Method      string                       `json:"method" enums:"GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE" example:"GET"`
	Headers     map[string]FetchHeaderValues `json:"headers"`
	Body        string                       `json:"body" format:"byte" example:""`
	Type        string                       `json:"type" enums:"Manual,Scheduled" example:"Manual"`
} // @name FetchRequest

type FetchResponse struct {
	Attempt int64 `json:"attempt" example:"1"`
} // @name FetchResponse

type GatewayError struct {
	Code    int    `json:"code" example:"3"`
	Message string `json:"message" example:"invalid request body"`
} // @name GatewayError

type ModuleDTO struct {
	ID          int           `json:"id" example:"1"`
	OwnerID     int           `json:"owner_id" example:"7"`
	Name        string        `json:"name" example:"Production monitoring"`
	Description string        `json:"description" example:"Checks public endpoints"`
	Webhooks    []*WebhookDTO `json:"webhooks"`
} // @name ModuleDTO

type ModuleList struct {
	Modules []*ModuleDTO `json:"modules"`
} // @name ModuleList

type ModuleRequestResponse struct {
	Module ModuleDTO `json:"module"`
} // @name ModuleRequestResponse

type UserLoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"strong-password"`
} // @name UserLoginRequest

type UserLoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOi..."`
} // @name UserLoginResponse

type UserRegisterRequest struct {
	Password string `json:"password" example:"strong-password"`
	Name     string `json:"name" example:"Maksim"`
	Email    string `json:"email" example:"user@example.com"`
} // @name UserRegisterRequest

type UserRegisterResponse struct {
	Email string `json:"email" example:"user@example.com"`
	Name  string `json:"name" example:"Maksim"`
} // @name UserRegisterResponse

type WebhookDTO struct {
	ID          int                 `json:"id" example:"1"`
	ModuleID    int                 `json:"module_id" example:"1"`
	Description string              `json:"description" example:"Ping API"`
	Interval    string              `json:"interval" example:"1m0s"`
	Timeout     string              `json:"timeout" example:"5s"`
	URL         string              `json:"url" example:"https://example.com/health"`
	Method      string              `json:"method" enums:"GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE" example:"GET"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body" format:"byte" example:""`
} // @name WebhookDTO

type WebhookDTORequestResponse struct {
	Webhook *WebhookDTO `json:"webhook"`
} // @name WebhookDTORequestResponse

// @Summary		Получить аналитику
// @Description	Возвращает агрегированные метрики по выполненным HTTP-запросам: количество вызовов, успешность, статусы, длительность и попытки.
// @Tags		Аналитика
// @Produce		json
// @Success		200 {object} AnalyticsResponse
// @Failure		500 {string} string
// @Router		/analytics [get]
func analyticsEndpointDoc() {}

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
func loginEndpointDoc() {}

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
func registerEndpointDoc() {}

// @Summary		Выполнить HTTP-запрос вручную
// @Description	Запускает один HTTP-запрос по переданным параметрам вебхука через gRPC-gateway.
// @Tags		Запуск
// @Accept		json
// @Produce		json
// @Param		request body FetchRequest true "Параметры запроса"
// @Success		200 {object} FetchResponse
// @Failure		400 {object} GatewayError
// @Failure		500 {object} GatewayError
// @Router		/fetch [post]
func fetchEndpointDoc() {}

// @Summary		Создать модуль
// @Description	Создаёт модуль пользователя вместе со списком вложенных вебхуков.
// @Tags		Модули
// @Accept		json
// @Produce		json
// @Param		request body ModuleRequestResponse true "Данные модуля"
// @Success		201 {object} ModuleRequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/ [post]
// @Security	Bearer
func createModuleEndpointDoc() {}

// @Summary		Обновить модуль
// @Description	Полностью обновляет данные модуля. Модуль должен принадлежать текущему пользователю.
// @Tags		Модули
// @Accept		json
// @Produce		json
// @Param		request body ModuleRequestResponse true "Данные модуля"
// @Success		200 {object} ModuleRequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		403 {object} APIError
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/ [put]
// @Security	Bearer
func updateModuleEndpointDoc() {}

// @Summary		Удалить модуль
// @Description	Удаляет модуль текущего пользователя по идентификатору.
// @Tags		Модули
// @Produce		json
// @Param		id path int true "ID модуля"
// @Success		204
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{id} [delete]
// @Security	Bearer
func deleteModuleEndpointDoc() {}

// @Summary		Получить модуль
// @Description	Возвращает один модуль текущего пользователя вместе с его вебхуками.
// @Tags		Модули
// @Produce		json
// @Param		id path int true "ID модуля"
// @Success		200 {object} ModuleRequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{id} [get]
// @Security	Bearer
func getModuleEndpointDoc() {}

// @Summary		Получить список модулей
// @Description	Возвращает все модули текущего пользователя.
// @Tags		Модули
// @Produce		json
// @Success		200 {object} ModuleList
// @Failure		401 {string} string
// @Failure		500 {object} APIError
// @Router		/modules/ [get]
// @Security	Bearer
func listModulesEndpointDoc() {}

// @Summary		Создать вебхук
// @Description	Добавляет вебхук в модуль текущего пользователя.
// @Tags		Вебхуки
// @Accept		json
// @Produce		json
// @Param		module_id path int true "ID модуля"
// @Param		request body WebhookDTORequestResponse true "Данные вебхука"
// @Success		201 {object} WebhookDTORequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{module_id}/webhook/ [post]
// @Security	Bearer
func createWebhookEndpointDoc() {}

// @Summary		Обновить вебхук
// @Description	Полностью обновляет вебхук внутри модуля текущего пользователя.
// @Tags		Вебхуки
// @Accept		json
// @Produce		json
// @Param		module_id path int true "ID модуля"
// @Param		request body WebhookDTORequestResponse true "Данные вебхука"
// @Success		200 {object} WebhookDTORequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{module_id}/webhook/ [put]
// @Security	Bearer
func updateWebhookEndpointDoc() {}

// @Summary		Удалить вебхук
// @Description	Удаляет вебхук из модуля текущего пользователя.
// @Tags		Вебхуки
// @Accept		json
// @Produce		json
// @Param		module_id path int true "ID модуля"
// @Param		webhook_id path int true "ID вебхука"
// @Param		request body WebhookDTORequestResponse true "Данные вебхука"
// @Success		204
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		409 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{module_id}/webhook/{webhook_id} [delete]
// @Security	Bearer
func deleteWebhookEndpointDoc() {}
