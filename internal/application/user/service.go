package user

import (
	"context"
	"fmt"
	"time"

	"mltestsuite/internal/domain/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct{ repo user.Repository }

func NewService(repo user.Repository) *Service { return &Service{repo: repo} }

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context) ([]*user.User, error) {
	return s.repo.FindAll(ctx)
}

type UpdateUserInput struct {
	Name     string
	Email    string
	Role     user.Role
	Active   bool
	Password string
	TeamID   *uuid.UUID
}

func (s *Service) CreateUser(ctx context.Context, input UpdateUserInput) error {
	if input.Password == "" {
		return fmt.Errorf("la contraseña es obligatoria")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hasheando contraseña: %w", err)
	}
	u := &user.User{
		ID:           uuid.New(),
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         input.Role,
		Active:       true,
		TeamID:       input.TeamID,
		CreatedAt:    time.Now(),
	}
	return s.repo.Save(ctx, u)
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.Name = input.Name
	u.Email = input.Email
	u.Role = input.Role
	u.Active = input.Active
	u.TeamID = input.TeamID
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("error hasheando contraseña: %w", err)
		}
		u.PasswordHash = string(hash)
	}
	return s.repo.Update(ctx, u)
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) UpdateNotificationEmails(ctx context.Context, id uuid.UUID, emails string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.NotificationEmails = emails
	return s.repo.Update(ctx, u)
}
