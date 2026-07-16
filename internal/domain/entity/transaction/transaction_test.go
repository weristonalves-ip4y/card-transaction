package transaction

import (
	"testing"

	"card-transaction/internal/domain/vo"
)

func TestToPersistenceMapKeepsNilAccountAndCardWhenUnresolved(t *testing.T) {
	tx, err := New("purchase-1", TypePurchase, ProductCard, vo.MustFromCents(100), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := tx.ToPersistenceMap()

	account, ok := payload["account_id"].(*int64)
	if !ok {
		t.Fatalf("expected account_id to be *int64, got %#v", payload["account_id"])
	}
	if account != nil {
		t.Fatalf("expected account_id nil, got %v", payload["account_id"])
	}

	cardID, ok := payload["card_id"].(*int64)
	if !ok {
		t.Fatalf("expected card_id to be *int64, got %#v", payload["card_id"])
	}
	if cardID != nil {
		t.Fatalf("expected card_id nil, got %v", payload["card_id"])
	}
}

func TestToPersistenceMapUsesResolvedAccountAndCard(t *testing.T) {
	tx, err := New("purchase-2", TypePurchase, ProductCard, vo.MustFromCents(100), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx = tx.WithResolvedAccountCard(10, 20)
	payload := tx.ToPersistenceMap()

	account, ok := payload["account_id"].(*int64)
	if !ok || account == nil || *account != 10 {
		t.Fatalf("expected account_id pointer to 10, got %#v", payload["account_id"])
	}

	cardID, ok := payload["card_id"].(*int64)
	if !ok || cardID == nil || *cardID != 20 {
		t.Fatalf("expected card_id pointer to 20, got %#v", payload["card_id"])
	}
}
