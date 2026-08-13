package domain

type Advisor struct {
	Id            string
	TaxRegime     string
	Debts         []Debt
	Incomes       []Income
	Expenses      float64
	Limits        []Limit
	Notifications []string
}

type Income struct {
	Id     string
	Amount float64
	Type   string
	Date   string
}

func (a *Advisor) TotalIncome() float64 {
	total := 0.0
	for _, income := range a.Incomes {
		total += income.Amount
	}
	return total
}

func (a *Advisor) TotalDebt() float64 {
	total := 0.0
	for _, deb := range a.Debts {
		total += deb.Balance
	}
	return total
}

func (a *Advisor) TotalDebtService() float64 {
	total := 0.0
	for _, deb := range a.Debts {
		total += deb.Installment
	}
	return total
}

func (a *Advisor) PaymentCapacity() float64 {
	return a.TotalIncome() - a.Expenses - a.TotalDebtService()
}
