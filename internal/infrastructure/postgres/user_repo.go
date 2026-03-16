package postgres

import (
	"context"
	"errors"
	"time"

	"mltestsuite/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) *UserRepository { return &UserRepository{pool: pool} }

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	var teamID interface{}
	if u.TeamID != nil {
		teamID = *u.TeamID
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, role, active, notification_emails, team_id, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		u.ID, u.Name, u.Email, u.PasswordHash, string(u.Role), u.Active, u.NotificationEmails, teamID, u.CreatedAt)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT u.id,u.name,u.email,u.password_hash,u.role,u.active,u.notification_emails,u.team_id,u.created_at,COALESCE(t.name,'')
		 FROM users u LEFT JOIN teams t ON t.id=u.team_id WHERE u.id=$1`, id)
	return scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT u.id,u.name,u.email,u.password_hash,u.role,u.active,u.notification_emails,u.team_id,u.created_at,COALESCE(t.name,'')
		 FROM users u LEFT JOIN teams t ON t.id=u.team_id WHERE u.email=$1`, email)
	return scanUser(row)
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*user.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id,u.name,u.email,u.password_hash,u.role,u.active,u.notification_emails,u.team_id,u.created_at,COALESCE(t.name,'')
		 FROM users u LEFT JOIN teams t ON t.id=u.team_id ORDER BY u.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*user.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	var teamID interface{}
	if u.TeamID != nil {
		teamID = *u.TeamID
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET name=$2,email=$3,password_hash=$4,role=$5,active=$6,notification_emails=$7,team_id=$8 WHERE id=$1`,
		u.ID, u.Name, u.Email, u.PasswordHash, string(u.Role), u.Active, u.NotificationEmails, teamID)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func scanUser(row pgx.Row) (*user.User, error) {
	var u user.User
	var role string
	var createdAt time.Time
	var teamID pgtype.UUID
	var teamName string
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Active, &u.NotificationEmails, &teamID, &createdAt, &teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	u.Role = user.Role(role)
	u.CreatedAt = createdAt
	u.TeamName = teamName
	if teamID.Valid {
		id := uuid.UUID(teamID.Bytes)
		u.TeamID = &id
	}
	return &u, nil
}

func scanUserFromRows(rows pgx.Rows) (*user.User, error) {
	var u user.User
	var role string
	var createdAt time.Time
	var teamID pgtype.UUID
	var teamName string
	if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Active, &u.NotificationEmails, &teamID, &createdAt, &teamName); err != nil {
		return nil, err
	}
	u.Role = user.Role(role)
	u.CreatedAt = createdAt
	u.TeamName = teamName
	if teamID.Valid {
		id := uuid.UUID(teamID.Bytes)
		u.TeamID = &id
	}
	return &u, nil
}
