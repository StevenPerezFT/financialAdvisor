package pgrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"financialAdvisor/services/core/internal/finance/domain"
)

type MovementRepo struct {
	pool *pgxpool.Pool
}

func NewMovementRepo(pool *pgxpool.Pool) *MovementRepo {
	return &MovementRepo{pool: pool}
}

func (m *MovementRepo) Create(ctx context.Context, userID string, movement domain.Movement) (domain.Movement, error) {
	err := m.pool.QueryRow(ctx,
		`INSERT INTO movements (user_id, type, amount, category, description) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, movement.Type, movement.Amount, movement.Category, movement.Description,
	).Scan(&movement.Id)

	return movement, err
}
