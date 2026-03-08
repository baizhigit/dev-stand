package convert

import (
	"github.com/shopspring/decimal"
	"google.golang.org/genproto/googleapis/type/money"
)

const nanosPerUnit = 1_000_000_000 // named constant, not magic number "1e9"

func MoneyToDecimal(m *money.Money) decimal.Decimal {
	if m == nil {
		return decimal.Zero // nil guard — proto fields are always potentially nil
	}
	units := decimal.NewFromInt(m.Units)
	nanos := decimal.NewFromInt32(m.Nanos)
	return units.Add(nanos.Div(decimal.NewFromInt(nanosPerUnit)))
}

func DecimalToMoney(d decimal.Decimal, currencyCode string) *money.Money {
	units := d.IntPart()
	// Correct: get fractional part first, then scale to nanos
	nanos := d.Sub(decimal.NewFromInt(units)).
		Mul(decimal.NewFromInt(nanosPerUnit)).
		IntPart()
	return &money.Money{
		CurrencyCode: currencyCode, // never omit — Money without currency is incomplete
		Units:        units,
		Nanos:        int32(nanos),
	}
}
