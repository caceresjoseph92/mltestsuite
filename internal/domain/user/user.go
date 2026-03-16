package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

var (
	ErrNotFound    = errors.New("usuario no encontrado")
	ErrEmailTaken  = errors.New("el email ya esta registrado")
	ErrInvalidRole = errors.New("rol invalido")
)

type User struct {
	ID                 uuid.UUID
	Name               string
	Email              string
	PasswordHash       string
	Role               Role
	Active             bool
	NotificationEmails string
	CreatedAt          time.Time
}

func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
}
