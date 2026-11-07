package tasks_transport_http

import (
	"fmt"
	"net/http"

	core_logger "github.com/vanya-9/golangproject/internal/core/logger"
	core_http_request "github.com/vanya-9/golangproject/internal/core/transport/http/request"
	core_http_response "github.com/vanya-9/golangproject/internal/core/transport/http/response"
)

type GetTasksResponse []TaskDTOResponse

type TaskResponse struct {
	ID           int     `json:"id"`
	Version      int     `json:"version"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Completed    bool    `json:"completed"`
	CreatedAt    string  `json:"created_at"`
	CompletedAt  *string `json:"completed_at"`
	AuthorUserID int     `json:"author_user_id"`
}

func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, limit, offset, err := getUserIDLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get query params",
		)

		return
	}

	tasksDomains, err := h.tasksService.GetTasks(ctx, userID, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get tasks",
		)

		return
	}

	response := GetTasksResponse(taskDTOsFromDomains(tasksDomains))

	responseHandler.JSONResponse(response, http.StatusOK)

}

func getUserIDLimitOffsetQueryParams(r *http.Request) (*int, *int, *int, error) {

	const (
		userIDQueryParamKey = "user_id"
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)

	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'user_id' query param: %w", err)
	}

	limit, err := core_http_request.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'limit' query param: %w", err)
	}

	offset, err := core_http_request.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'offset' query param: %w", err)
	}

	return userID, limit, offset, nil
}
