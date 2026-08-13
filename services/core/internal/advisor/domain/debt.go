package domain

type Debt struct {
	Id          string
	Creditor    string
	Balance     float64
	Installment float64
	DueDate     string
}
