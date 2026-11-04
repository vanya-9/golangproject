package user_transport_http

import (
	"context"
	"net/http"

	"github.com/vanya-9/golangproject/internal/core/domain"
	core_http_server "github.com/vanya-9/golangproject/internal/core/transport/http/server"
)

type UsersHTTPHandler struct { //будет вызывать слой сервиса,
	// зависит от интерфейса для облегчения тестирования и гибкости:
	//  можно будет описать новый сервис и благодаря зависимости через интерфейс
	// легко их переопределять
	usersService UsersService
}

type UsersService interface { //интерфейс сервиса
	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)

	GetUsers(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.User, error)

	GetUser(
		ctx context.Context,
		id int,
	) (domain.User, error)

	DeleteUser(
		ctx context.Context,
		id int,
	) error

	PatchUser(
		ctx context.Context,
		id int,
		patch domain.UserPatch,
	) (domain.User, error)
}

func NewUsersHTTPHandler(usersService UsersService) *UsersHTTPHandler { //конструктор
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}

func (h *UsersHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/users",
			Handler: h.CreateUser,
		},
		{
			Method:  http.MethodGet,
			Path:    "/users",
			Handler: h.GetUsers,
			// Middleware: []core_http_middleware.Middleware{
			// 	core_http_middleware.Dummy("get users middleware"),
			// },
		},
		{
			Method:  http.MethodGet,
			Path:    "/users/{id}",
			Handler: h.GetUser,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/{id}",
			Handler: h.DeleteUser,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/{id}",
			Handler: h.PatchUser,
		},
	}
}
