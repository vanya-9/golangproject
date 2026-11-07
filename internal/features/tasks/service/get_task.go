package tasks_service

import (
	"context"

	"github.com/vanya-9/golangproject/internal/core/domain"
)

func (s *TasksService) GetTask(
	ctx context.Context,
	id int,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	return task, nil
}
