package usecase

import (
	"financialAdvisor/services/core/internal/advisor/domain"
	"financialAdvisor/services/core/internal/const/advisor/taxes"
	"fmt"
)

type AdvisorUseCase struct{}

func NewAdvisorUseCase() *AdvisorUseCase {
	return &AdvisorUseCase{}
}

func (u *AdvisorUseCase) RegisterIncome(advisor *domain.Advisor, incomes []domain.Income) {
	advisor.Incomes = append(advisor.Incomes, incomes...)
}

func (u *AdvisorUseCase) RegisterDebt(advisor *domain.Advisor, debts []domain.Debt) {
	advisor.Debts = append(advisor.Debts, debts...)
}

func (u *AdvisorUseCase) CalculateTax(advisor *domain.Advisor) float64 {
	total := advisor.TotalIncome()
	switch advisor.TaxRegime {
	case taxes.SmallContributor:
		return total * 0.05
	case taxes.OptionalSimplified:
		return total * 0.07
	case taxes.ProfitsTax:
		return total * 0.25
	default:
		return 0
	}
}

func (u *AdvisorUseCase) RegisterExpense(advisor *domain.Advisor, amount float64, category string) {
	advisor.Expenses += amount

	for _, limit := range advisor.Limits {
		if limit.Category == category && advisor.Expenses >= limit.Value {
			advisor.Notifications = append(advisor.Notifications, fmt.Sprintf("Limit reached for %s", category))
		}
	}
}
