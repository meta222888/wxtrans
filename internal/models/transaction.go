package models

import "time"

type Transaction struct {
	ID            int64
	TransTime     time.Time
	TransType     string
	Counterparty  string
	Product       string
	Direction     string
	Amount        float64
	PaymentMethod string
	Status        string
	TransNo       string
	MerchantNo    string
	Remark        string
	ImportedAt    time.Time
}

type SearchFilter struct {
	Keyword   string
	DateFrom  *time.Time
	DateTo    *time.Time
	Direction string
	TransType string
	AmountMin float64 // 0 表示不限
	AmountMax float64 // 0 表示不限
	Limit     int
	Offset    int
}

type Summary struct {
	TotalCount    int
	IncomeCount   int
	ExpenseCount  int
	IncomeAmount  float64
	ExpenseAmount float64
	NetAmount     float64
}

type TypeSummary struct {
	TransType string
	Direction string
	Count     int
	Amount    float64
}

type MonthSummary struct {
	Month         string
	IncomeAmount  float64
	ExpenseAmount float64
	NetAmount     float64
}

type ImportResult struct {
	Imported int
	Skipped  int
	Errors   int
	Message  string
}
