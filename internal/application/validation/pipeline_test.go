package validation

import (
	"testing"

	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/transaction"
	"card-transaction/internal/domain/vo/money"
)

func TestPipelineStopsOnDuplicate(t *testing.T) {
	tx, err := transaction.New(
		"id",
		"identifier",
		"acc",
		transaction.TypePurchase,
		transaction.ProductCard,
		money.MustFromCents(100),
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := Validate(Context{
		Transaction: tx,
		Card: card.Card{
			ID:           "card-1",
			AccountID:    "acc",
			Status:       card.StatusActive,
			ProductCode:  "011202",
			CheckLimit:   true,
			MonthlyLimit: money.MustFromCents(1000),
		},
		IsDuplicate:           true,
		CardProductCode:       "011202",
		MonthlyTransactionSum: money.MustFromCents(0),
		Balance:               money.MustFromCents(1000),
	})

	if result.Approved {
		t.Fatalf("duplicate transaction should be rejected")
	}
	if result.Code != "07" {
		t.Fatalf("expected code 07, got %s", result.Code)
	}
}

func TestPipelineApprovesHappyPath(t *testing.T) {
	tx, err := transaction.New(
		"id",
		"identifier",
		"acc",
		transaction.TypePurchase,
		transaction.ProductCard,
		money.MustFromCents(100),
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := Validate(Context{
		Transaction: tx,
		Card: card.Card{
			ID:           "card-1",
			AccountID:    "acc",
			Status:       card.StatusActive,
			ProductCode:  "011202",
			CheckLimit:   true,
			MonthlyLimit: money.MustFromCents(1000),
		},
		IsDuplicate:           false,
		CardProductCode:       "011202",
		MonthlyTransactionSum: money.MustFromCents(100),
		Balance:               money.MustFromCents(900),
	})

	if !result.Approved {
		t.Fatalf("expected approved result, got code=%s", result.Code)
	}
	if result.Code != "00" {
		t.Fatalf("expected code 00, got %s", result.Code)
	}
}
