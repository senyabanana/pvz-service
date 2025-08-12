package repository

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/senyabanana/pvz-service/internal/entity"
)

type UserPostgres struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewUserPostgres(db *sqlx.DB) *UserPostgres {
	return &UserPostgres{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *UserPostgres) CreateUser(ctx context.Context, user *entity.User) error {
	user.ID = uuid.New()
	query := `INSERT INTO users (id, email, password_hash, role, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).
		ExecContext(ctx, query, user.ID, user.Email, user.Password, user.Role, user.CreatedAt)

	return err
}

func (r *UserPostgres) IsEmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &count, query, email)

	return count > 0, err
}

func (r *UserPostgres) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
