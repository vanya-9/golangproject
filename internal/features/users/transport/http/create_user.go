package user_transport_http

import (
	"net/http"

	"github.com/vanya-9/golangproject/internal/core/domain"
	core_logger "github.com/vanya-9/golangproject/internal/core/logger"
	core_http_request "github.com/vanya-9/golangproject/internal/core/transport/http/request"
	core_http_response "github.com/vanya-9/golangproject/internal/core/transport/http/response"
)

// дтошка, описывает запрос на создание пользователя \
// мы сами на уровне сервиса или бд выдадим id, поэтому тут нет id, а в ответе есть
type CreateUserRequest struct {
	FullName    string  `json:"full_name" validate:"required,min=3,max=100"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,min=10,max=15,startswith=+"`
}

// дтошка, описывает ответ на запрос создания пользователя
type CreateUserResponse UserDTOResponse

// метод HTTPHandler транспорта. Обработчик запроса на создание пользователя
func (h *UsersHTTPHandler) CreateUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request CreateUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate http request")
		return
	}

	userDomain := domainFromDTO(request)

	userDomain, err := h.usersService.CreateUser(ctx, userDomain)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create user")

		return
	}

	response := CreateUserResponse(userDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}

func domainFromDTO(dto CreateUserRequest) domain.User {
	return domain.NewUserUninitialized(dto.FullName, dto.PhoneNumber)
}
