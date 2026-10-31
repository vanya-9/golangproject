package core_http_server

import "net/http"

// Route — это описание одного конкретного HTTP-эндпоинта.
// Роль — задать правило: «на такой-то метод и путь отвечай таким обработчиком».
// Это атомарная единица маршрутизации.
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

func NewRoute(method string, path string, handler http.HandlerFunc) Route {
	return Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}
