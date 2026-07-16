package vo

import "fmt"

// Money represents an amount in cents to avoid floating-point errors.
type Money struct {
	cents int64
}

func NewFromCents(cents int64) (Money, error) {
	if cents < 0 {
		return Money{}, fmt.Errorf("money cannot be negative: %d", cents)
	}
	return Money{cents: cents}, nil
}

func MustFromCents(cents int64) Money {
	m, err := NewFromCents(cents)
	if err != nil {
		panic(err)
	}
	return m
}

func Zero() Money {
	return Money{cents: 0}
}

func (m Money) ToFloat() float64 {
	return float64(m.cents) / 100
}

func (m Money) Cents() int64 {
	return m.cents
}

func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.cents < other.cents {
		return Money{}, fmt.Errorf("insufficient amount: have=%d need=%d", m.cents, other.cents)
	}
	return Money{cents: m.cents - other.cents}, nil
}

func (m Money) LessThan(other Money) bool {
	return m.cents < other.cents
}

func (m Money) IsZero() bool {
	return m.cents == 0
}

func (m Money) GreaterThan(other Money) bool {
	return m.cents > other.cents
}
