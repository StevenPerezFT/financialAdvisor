package pgrepo

import (
	"context"
	"financialAdvisor/services/core/internal/advisor/domain"
	"financialAdvisor/services/core/internal/const/advisor/dateformat"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DebtRepo struct {
	pool *pgxpool.Pool
}

func NewDebtRepo(pool *pgxpool.Pool) *DebtRepo {
	return &DebtRepo{pool: pool}
}
func (d *DebtRepo) Create(ctx context.Context, userID string, debt domain.Debt) (domain.Debt, error) {
	dueDate, err := time.Parse(dateformat.Layout, debt.DueDate)
	if err != nil {
		return domain.Debt{}, err
	}

	err = d.pool.QueryRow(ctx, `INSERT INTO debts (user_id, creditor, balance, installment, due_date) 
	VALUES ($1, $2, $3, $4, $5) RETURNING id`, userID, debt.Creditor, debt.Balance, debt.Installment, dueDate,
	).Scan(&debt.Id)

	if err != nil {
		return domain.Debt{}, err
	}
	return debt, nil
}
