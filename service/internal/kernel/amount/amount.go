package amount

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var ErrDivisionByZero = errors.New("division by zero")

type Amount decimal.Decimal

func FromString(value string) (Amount, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Amount{}, err
	}
	return Amount(d), nil
}

func FromFloat64(value float64) Amount {
	return Amount(decimal.NewFromFloat(value))
}

func (a Amount) Value() (driver.Value, error) {
	return decimal.Decimal(a).String(), nil
}

func (a *Amount) Scan(value any) error {
	if value == nil {
		*a = Zero()
		return nil
	}
	switch typed := value.(type) {
	case []byte:
		return a.scanString(string(typed))
	case string:
		return a.scanString(typed)
	case float64:
		*a = FromFloat64(typed)
		return nil
	case int64:
		*a = FromInt64(typed)
		return nil
	default:
		return fmt.Errorf("unsupported type %T for amount", value)
	}
}

func (a *Amount) scanString(value string) error {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return fmt.Errorf("invalid numeric value %q: %w", value, err)
	}
	*a = Amount(parsed)
	return nil
}

func FromInt64(value int64) Amount {
	return Amount(decimal.New(value, 0))
}

func Zero() Amount {
	return Amount(decimal.Zero)
}

func (a Amount) Add(b Amount) Amount {
	return Amount(decimal.Decimal(a).Add(decimal.Decimal(b)))
}

func (a Amount) Sub(b Amount) Amount {
	return Amount(decimal.Decimal(a).Sub(decimal.Decimal(b)))
}

func (a Amount) Mul(b Amount) Amount {
	return Amount(decimal.Decimal(a).Mul(decimal.Decimal(b)))
}

func (a Amount) Div(b Amount) (Amount, error) {
	if b.IsZero() {
		return Amount{}, ErrDivisionByZero
	}
	return Amount(decimal.Decimal(a).DivRound(decimal.Decimal(b), 18)), nil
}

func (a Amount) Round(dp int32) Amount {
	return Amount(decimal.Decimal(a).Round(dp))
}

func (a Amount) Neg() Amount {
	return Amount(decimal.Decimal(a).Neg())
}

func (a Amount) Abs() Amount {
	return Amount(decimal.Decimal(a).Abs())
}

func (a Amount) Sign() int {
	return decimal.Decimal(a).Sign()
}

func (a Amount) IsZero() bool {
	return decimal.Decimal(a).IsZero()
}

func (a Amount) IsNegative() bool {
	return decimal.Decimal(a).IsNegative()
}

func (a Amount) IsPositive() bool {
	return decimal.Decimal(a).IsPositive()
}

func (a Amount) Equal(b Amount) bool {
	return decimal.Decimal(a).Equal(decimal.Decimal(b))
}

func (a Amount) GreaterThan(b Amount) bool {
	return decimal.Decimal(a).GreaterThan(decimal.Decimal(b))
}

func (a Amount) LessThan(b Amount) bool {
	return decimal.Decimal(a).LessThan(decimal.Decimal(b))
}

func (a Amount) String() string {
	return decimal.Decimal(a).String()
}

func (a Amount) Float64() float64 {
	return decimal.Decimal(a).InexactFloat64()
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(decimal.Decimal(a).String()), nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var value decimal.Decimal
	if string(data) == "null" {
		*a = Zero()
		return nil
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*a = Amount(value)
	return nil
}

func (a Amount) Int64() int64 {
	return decimal.Decimal(a).IntPart()
}

func Net(amounts []Amount, dp int32) Amount {
	total := decimal.Zero
	for _, a := range amounts {
		total = total.Add(decimal.Decimal(a))
	}
	return Amount(total).Round(dp)
}

func IsBalanced(debits, credits Amount, dp int32) bool {
	return debits.Round(dp).Equal(credits.Round(dp))
}
