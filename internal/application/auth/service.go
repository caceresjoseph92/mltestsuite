package auth

import (
	"context"
	"errors"
	"time"

	"mltestsuite/internal/domain/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("credenciales inválidas")

type Service struct{ repo user.Repository }

func NewService(repo user.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, user.ErrEmailTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &user.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         user.RoleUser,
		Active:       true,
		CreatedAt:    time.Now(),
	}
	// First user becomes admin
	count, _ := s.repo.Count(ctx)
	if count == 0 {
		u.Role = user.RoleAdmin
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*user.User, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !u.Active {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}
