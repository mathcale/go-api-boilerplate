package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
)

type User interface {
	Exists(ctx context.Context, email string) (*bool, error)
	Get(ctx context.Context, email string) (*models.User, error)
	Save(ctx context.Context, user models.User) error
}

type user struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) User {
	return &user{
		db: db,
	}
}

func (r *user) Exists(ctx context.Context, email string) (*bool, error) {
	var exists bool

	if err := r.db.GetContext(ctx, &exists, queryUserExists, email); err != nil {
		return nil, err
	}

	return &exists, nil
}

func (r *user) Get(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	if err := r.db.GetContext(ctx, &user, queryGetUserByEmail, email); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *user) Save(ctx context.Context, u models.User) error {
	_, err := r.db.ExecContext(ctx, queryInsertUser, u.Name, u.Email, u.Password, u.Active)
	if err != nil {
		return err
	}

	return nil
}
