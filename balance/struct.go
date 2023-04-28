package balance

import "github.com/shopspring/decimal"

type Asset struct {
	Symbol string
	Amount decimal.Decimal
	Value  decimal.Decimal
}
