package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/vanya-9/golangproject/internal/core/transport/http/middleware"
)

// Route — это описание одного конкретного HTTP-эндпоинта.
// Роль — задать правило: «на такой-то метод и путь отвечай таким обработчиком».
// Это атомарная единица маршрутизации.
type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []core_http_middleware.Middleware
}

func (r *Route) WithMiddleware() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r.Handler,
		r.Middleware...,
	)
}
