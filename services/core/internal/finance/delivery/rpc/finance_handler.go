package rpc

import (
	"context"

	"financialAdvisor/services/core/internal/finance/domain"
	"financialAdvisor/services/core/internal/finance/infra/pgrepo"
	financev1 "financialAdvisor/services/core/internal/pb/financialadvisor/finance/v1"
)

type FinanceHandler struct {
	financev1.UnimplementedFinanceServiceServer
	movementRepo *pgrepo.MovementRepo
}

func NewFinanceHandler(movementRepo *pgrepo.MovementRepo) *FinanceHandler {
	return &FinanceHandler{movementRepo: movementRepo}
}

func (f *FinanceHandler) RegisterMovement(ctx context.Context, req *financev1.RegisterMovementRequest) (*financev1.RegisterMovementResponse, error) {
	movement := domain.Movement{
		Type:        req.Movement.Type,
		Amount:      req.Movement.Amount,
		Category:    req.Movement.Category,
		Description: req.Movement.Description,
	}

	saved, err := f.movementRepo.Create(ctx, req.UserId, movement)
	if err != nil {
		return nil, err
	}

	return &financev1.RegisterMovementResponse{
		Movement: &financev1.Movement{
			Id:          saved.Id,
			Type:        saved.Type,
			Amount:      saved.Amount,
			Category:    saved.Category,
			Description: saved.Description,
		},
	}, nil
}
