package service

import (
	"context"
	"errors"
	"strings"

	"bico/internal/domain"
	"bico/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
	// Regras de validação
	if strings.TrimSpace(user.Nome) == "" {
		return errors.New("o campo 'nome' é obrigatório")
	}
	if strings.TrimSpace(user.Email) == "" || !strings.Contains(user.Email, "@") {
		return errors.New("forneça um 'email' válido")
	}

	return s.repo.Create(ctx, user)
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("o ID do usuário é obrigatório")
	}
	return s.repo.GetByID(ctx, id)
}
