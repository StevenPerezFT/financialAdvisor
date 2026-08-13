package domain

import "testing"

func TestTotalIncome(t *testing.T) {
	tests := []struct {
		name    string
		incomes []Income
		want    float64
	}{
		{"no incomes", nil, 0},
		{"single income", []Income{{Amount: 1200}}, 1200},
		{"multiple incomes", []Income{{Amount: 3000}, {Amount: 500}, {Amount: 250.50}}, 3750.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advisor := &Advisor{Incomes: tt.incomes}
			if got := advisor.TotalIncome(); got != tt.want {
				t.Errorf("TotalIncome() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTotalDebt(t *testing.T) {
	tests := []struct {
		name  string
		debts []Debt
		want  float64
	}{
		{"no debts", nil, 0},
		{"single debt", []Debt{{Balance: 5000}}, 5000},
		{"multiple debts", []Debt{{Balance: 5000}, {Balance: 1200.75}}, 6200.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advisor := &Advisor{Debts: tt.debts}
			if got := advisor.TotalDebt(); got != tt.want {
				t.Errorf("TotalDebt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTotalDebtService(t *testing.T) {
	tests := []struct {
		name  string
		debts []Debt
		want  float64
	}{
		{"no debts", nil, 0},
		{"single debt", []Debt{{Balance: 50000, Installment: 300}}, 300},
		{"multiple debts", []Debt{{Installment: 300}, {Installment: 200}}, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advisor := &Advisor{Debts: tt.debts}
			if got := advisor.TotalDebtService(); got != tt.want {
				t.Errorf("TotalDebtService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaymentCapacity(t *testing.T) {
	tests := []struct {
		name    string
		advisor *Advisor
		want    float64
	}{
		{
			name:    "income minus expenses minus installments",
			advisor: &Advisor{Incomes: []Income{{Amount: 3000}, {Amount: 500}}, Expenses: 1000, Debts: []Debt{{Installment: 300}, {Installment: 200}}},
			want:    2000,
		},
		{
			name:    "no debts",
			advisor: &Advisor{Incomes: []Income{{Amount: 2000}}, Expenses: 500},
			want:    1500,
		},
		{
			name:    "installments exceed income minus expenses",
			advisor: &Advisor{Incomes: []Income{{Amount: 1000}}, Expenses: 800, Debts: []Debt{{Installment: 500}}},
			want:    -300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.advisor.PaymentCapacity(); got != tt.want {
				t.Errorf("PaymentCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}
