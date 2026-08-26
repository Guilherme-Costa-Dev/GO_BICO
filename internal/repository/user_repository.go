package repository

import (
	"context"
	"fmt"

	"bico/internal/domain"
	"cloud.google.com/go/firestore"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type userRepository struct {
	client *firestore.Client
}

func NewUserRepository(client *firestore.Client) UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// Insere na coleção "usuarios". O Firebase criará a coleção se ela não existir!
	docRef, _, err := r.client.Collection("usuarios").Add(ctx, map[string]interface{}{
		"nome":  user.Nome,
		"email": user.Email,
	})
	if err != nil {
		return fmt.Errorf("erro ao salvar no firestore: %w", err)
	}

	user.ID = docRef.ID
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	doc, err := r.client.Collection("usuarios").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	var user domain.User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("erro ao mapear dados do usuário: %w", err)
	}

	user.ID = doc.Ref.ID
	return &user, nil
}
