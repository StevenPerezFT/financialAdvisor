package usecase

import "financialAdvisor/services/core/internal/finance/domain"

type MovementUseCase struct{}

func NewMovementUseCase() *MovementUseCase {
	return &MovementUseCase{}
}

func (u *MovementUseCase) RegisterMovement(movements *[]domain.Movement, movement domain.Movement) {
	*movements = append(*movements, movement)
}
