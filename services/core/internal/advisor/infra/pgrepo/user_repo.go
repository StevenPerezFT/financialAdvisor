package pgrepo

import (
	"context"
	"financialAdvisor/services/core/internal/advisor/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (u *UserRepo) Create(ctx context.Context, user domain.User) (domain.User, error) {

	err := u.pool.QueryRow(ctx,
		`INSERT INTO users (email, name, tax_regime) VALUES ($1, $2, $3) RETURNING id`,
		user.Email, user.Name, user.TaxRegime,
	).Scan(&user.Id)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}
