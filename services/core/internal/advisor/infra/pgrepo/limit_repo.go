package pgrepo

import (
	"context"
	"financialAdvisor/services/core/internal/advisor/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LimitRepo struct {
	pool *pgxpool.Pool
}

func NewLimitRepo(pool *pgxpool.Pool) *LimitRepo {
	return &LimitRepo{pool: pool}
}
func (l *LimitRepo) Create(ctx context.Context, userID string, limit domain.Limit) (domain.Limit, error) {
	err := l.pool.QueryRow(ctx,
		`INSERT INTO limits (user_id, category, value) VALUES ($1, $2, $3) RETURNING id`,
		userID, limit.Category, limit.Value,
	).Scan(&limit.Id)
	if err != nil {
		return domain.Limit{}, err
	}
	return limit, nil
}
