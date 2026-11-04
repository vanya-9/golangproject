package users_service

import (
	"context"
	"fmt"

	"github.com/vanya-9/golangproject/internal/core/domain"
)

func (s UsersService) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	//Валидируем домен
	//Сохраняем в репозиторий
	if err := user.Validate(); err != nil {
		return domain.User{}, fmt.Errorf("validate user domain: %w", err)
	}

	user, err := s.usersRepository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user in repository: %w", err)
	}

	return user, nil

}
