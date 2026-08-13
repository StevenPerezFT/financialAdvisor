package usecase

import (
	"testing"

	"financialAdvisor/services/core/internal/advisor/domain"
	"financialAdvisor/services/core/internal/const/advisor/taxes"
)

func TestCalculateTax(t *testing.T) {
	tests := []struct {
		name    string
		regime  string
		incomes []domain.Income
		want    float64
	}{
		{"small contributor", taxes.SmallContributor, []domain.Income{{Amount: 1000}}, 50},
		{"optional simplified", taxes.OptionalSimplified, []domain.Income{{Amount: 1000}}, 70},
		{"profits tax", taxes.ProfitsTax, []domain.Income{{Amount: 1000}}, 250},
		{"unknown regime", "Unknown", []domain.Income{{Amount: 1000}}, 0},
		{"no income", taxes.ProfitsTax, nil, 0},
	}

	u := NewAdvisorUseCase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advisor := &domain.Advisor{TaxRegime: tt.regime, Incomes: tt.incomes}
			if got := u.CalculateTax(advisor); got != tt.want {
				t.Errorf("CalculateTax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterIncome(t *testing.T) {
	u := NewAdvisorUseCase()
	advisor := &domain.Advisor{Incomes: []domain.Income{{Amount: 100}}}

	u.RegisterIncome(advisor, []domain.Income{{Amount: 200}, {Amount: 50}})

	if got, want := len(advisor.Incomes), 3; got != want {
		t.Fatalf("len(Incomes) = %v, want %v", got, want)
	}
	if got, want := advisor.TotalIncome(), 350.0; got != want {
		t.Errorf("TotalIncome() = %v, want %v", got, want)
	}
}

func TestRegisterDebt(t *testing.T) {
	u := NewAdvisorUseCase()
	advisor := &domain.Advisor{Debts: []domain.Debt{{Balance: 1000}}}

	u.RegisterDebt(advisor, []domain.Debt{{Balance: 500}, {Balance: 250}})

	if got, want := len(advisor.Debts), 3; got != want {
		t.Fatalf("len(Debts) = %v, want %v", got, want)
	}
	if got, want := advisor.TotalDebt(), 1750.0; got != want {
		t.Errorf("TotalDebt() = %v, want %v", got, want)
	}
}

func TestRegisterExpense(t *testing.T) {
	t.Run("notifies when limit is reached", func(t *testing.T) {
		u := NewAdvisorUseCase()
		advisor := &domain.Advisor{Limits: []domain.Limit{{Category: "Food", Value: 500}}}

		u.RegisterExpense(advisor, 500, "Food")

		if got, want := len(advisor.Notifications), 1; got != want {
			t.Fatalf("len(Notifications) = %v, want %v", got, want)
		}
	})

	t.Run("does not notify below limit", func(t *testing.T) {
		u := NewAdvisorUseCase()
		advisor := &domain.Advisor{Limits: []domain.Limit{{Category: "Food", Value: 500}}}

		u.RegisterExpense(advisor, 100, "Food")

		if got, want := len(advisor.Notifications), 0; got != want {
			t.Fatalf("len(Notifications) = %v, want %v", got, want)
		}
	})

	t.Run("does not notify for a different category", func(t *testing.T) {
		u := NewAdvisorUseCase()
		advisor := &domain.Advisor{Limits: []domain.Limit{{Category: "Food", Value: 500}}}

		u.RegisterExpense(advisor, 1000, "Transport")

		if got, want := len(advisor.Notifications), 0; got != want {
			t.Fatalf("len(Notifications) = %v, want %v", got, want)
		}
	})

	t.Run("accumulates expenses across calls", func(t *testing.T) {
		u := NewAdvisorUseCase()
		advisor := &domain.Advisor{}

		u.RegisterExpense(advisor, 300, "Food")
		u.RegisterExpense(advisor, 200, "Transport")

		if got, want := advisor.Expenses, 500.0; got != want {
			t.Errorf("Expenses = %v, want %v", got, want)
		}
	})
}
