package rpc

import (
	"context"

	"financialAdvisor/services/core/internal/advisor/domain"
	"financialAdvisor/services/core/internal/advisor/infra/pgrepo"
	"financialAdvisor/services/core/internal/advisor/usecase"
	advisorv1 "financialAdvisor/services/core/internal/pb/financialadvisor/advisor/v1"
)

type AdvisorHandler struct {
	advisorv1.UnimplementedAdvisorServiceServer
	userRepo   *pgrepo.UserRepo
	incomeRepo *pgrepo.IncomeRepo
	debtsRepo  *pgrepo.DebtRepo
	usecase    *usecase.AdvisorUseCase
}

func NewAdvisorHandler(userRepo *pgrepo.UserRepo, incomeRepo *pgrepo.IncomeRepo, debtsRepo *pgrepo.DebtRepo, usecase *usecase.AdvisorUseCase) *AdvisorHandler {
	return &AdvisorHandler{userRepo: userRepo, incomeRepo: incomeRepo, debtsRepo: debtsRepo, usecase: usecase}
}

func (h *AdvisorHandler) CalculateTax(ctx context.Context, req *advisorv1.CalculateTaxRequest) (*advisorv1.CalculateTaxResponse, error) {
	user, err := h.userRepo.FindByID(ctx, req.AdvisorId)
	if err != nil {
		return nil, err
	}

	incomes, err := h.incomeRepo.ListByUserID(ctx, req.AdvisorId)
	if err != nil {
		return nil, err
	}

	advisor := &domain.Advisor{TaxRegime: user.TaxRegime, Incomes: incomes}
	tax := h.usecase.CalculateTax(advisor)

	return &advisorv1.CalculateTaxResponse{Tax: tax}, nil
}

func (h *AdvisorHandler) RegisterIncomes(ctx context.Context, req *advisorv1.RegisterIncomesRequest) (*advisorv1.RegisterIncomesResponse, error) {
	created := make([]*advisorv1.Income, 0, len(req.Incomes))

	for _, protoIncome := range req.Incomes {
		income := domain.Income{
			Amount: protoIncome.Amount,
			Type:   protoIncome.Type,
			Date:   protoIncome.Date,
		}

		saved, err := h.incomeRepo.Create(ctx, req.AdvisorId, income)
		if err != nil {
			return nil, err
		}

		created = append(created, &advisorv1.Income{
			Id:     saved.Id,
			Amount: saved.Amount,
			Type:   saved.Type,
			Date:   saved.Date,
		})
	}

	return &advisorv1.RegisterIncomesResponse{Incomes: created}, nil
}

func (h *AdvisorHandler) RegisterDebts(ctx context.Context, req *advisorv1.RegisterDebtsRequest) (*advisorv1.RegisterDebtsResponse, error) {
	created := make([]*advisorv1.Debt, 0, len(req.Debts))

	for _, protoDebt := range req.Debts {
		debt := domain.Debt{
			Creditor:    protoDebt.Creditor,
			Balance:     protoDebt.Balance,
			Installment: protoDebt.Installment,
			DueDate:     protoDebt.DueDate,
		}

		saved, err := h.debtsRepo.Create(ctx, req.AdvisorId, debt)
		if err != nil {
			return nil, err
		}

		created = append(created, &advisorv1.Debt{
			Id:          saved.Id,
			Creditor:    saved.Creditor,
			Balance:     saved.Balance,
			Installment: saved.Installment,
			DueDate:     saved.DueDate,
		})
	}

	return &advisorv1.RegisterDebtsResponse{Debts: created}, nil
}
