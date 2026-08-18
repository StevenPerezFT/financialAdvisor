package pgrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"financialAdvisor/services/core/internal/advisor/domain"
	"financialAdvisor/services/core/internal/const/advisor/dateformat"
)

type IncomeRepo struct {
	pool *pgxpool.Pool
}

func NewIncomeRepo(pool *pgxpool.Pool) *IncomeRepo {
	return &IncomeRepo{pool: pool}
}

func (r *IncomeRepo) Create(ctx context.Context, userID string, income domain.Income) (domain.Income, error) {
	date, err := time.Parse(dateformat.Layout, income.Date)
	if err != nil {
		return domain.Income{}, err
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO incomes (user_id, amount, type, date) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, income.Amount, income.Type, date,
	).Scan(&income.Id)
	if err != nil {
		return domain.Income{}, err
	}
	return income, nil
}
