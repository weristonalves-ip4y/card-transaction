package sqlserver

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCardRepositoryFindByPaysmartIDHappyPath(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id",
		"account_id",
		"uuid",
		"card_id",
		"card_status_id",
		"card_monthly_limit",
		"card_check_limit",
		"ps_product_code",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		int64(10),
		int64(1522),
		"uuid-1",
		"pay-1",
		4,
		int64(100000),
		true,
		"011202",
		now,
		now,
		sql.NullTime{},
	)

	mock.ExpectQuery("SELECT TOP 1").WithArgs(sql.Named("paysmartID", "pay-1")).WillReturnRows(rows)

	repo := NewCardRepository(db)
	card, err := repo.FindByPaysmartID("pay-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if card.ID != 10 {
		t.Fatalf("expected card id 10, got %d", card.ID)
	}
	if card.AccountID != 1522 {
		t.Fatalf("expected account 1522, got %d", card.AccountID)
	}
	if card.CardMonthlyLimit.Cents() != 100000 {
		t.Fatalf("expected monthly limit 100000, got %d", card.CardMonthlyLimit.Cents())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCardRepositoryFindByPaysmartIDQueryError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT TOP 1").WithArgs(sql.Named("paysmartID", "pay-1")).WillReturnError(sql.ErrConnDone)

	repo := NewCardRepository(db)
	_, err = repo.FindByPaysmartID("pay-1")
	if err == nil {
		t.Fatalf("expected query error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCardRepositoryFindByPaysmartIDNullMonthlyLimit(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id",
		"account_id",
		"uuid",
		"card_id",
		"card_status_id",
		"card_monthly_limit",
		"card_check_limit",
		"ps_product_code",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		int64(11),
		int64(1523),
		"uuid-2",
		"pay-2",
		4,
		nil,
		true,
		"011202",
		now,
		now,
		sql.NullTime{},
	)

	mock.ExpectQuery("SELECT TOP 1").WithArgs(sql.Named("paysmartID", "pay-2")).WillReturnRows(rows)

	repo := NewCardRepository(db)
	card, err := repo.FindByPaysmartID("pay-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if card.CardMonthlyLimit.Cents() != 0 {
		t.Fatalf("expected monthly limit 0 when column is null, got %d", card.CardMonthlyLimit.Cents())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCardRepositoryFindByPaysmartIDNullCheckLimit(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id",
		"account_id",
		"uuid",
		"card_id",
		"card_status_id",
		"card_monthly_limit",
		"card_check_limit",
		"ps_product_code",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		int64(12),
		int64(1524),
		"uuid-3",
		"pay-3",
		4,
		int64(1500),
		nil,
		"011202",
		now,
		now,
		sql.NullTime{},
	)

	mock.ExpectQuery("SELECT TOP 1").WithArgs(sql.Named("paysmartID", "pay-3")).WillReturnRows(rows)

	repo := NewCardRepository(db)
	card, err := repo.FindByPaysmartID("pay-3")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if card.CardCheckLimit {
		t.Fatalf("expected card_check_limit false when column is null")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCardRepositoryFindByPaysmartIDNullProductCode(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id",
		"account_id",
		"uuid",
		"card_id",
		"card_status_id",
		"card_monthly_limit",
		"card_check_limit",
		"ps_product_code",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		int64(13),
		int64(1525),
		"uuid-4",
		"pay-4",
		4,
		int64(1000),
		true,
		nil,
		now,
		now,
		sql.NullTime{},
	)

	mock.ExpectQuery("SELECT TOP 1").WithArgs(sql.Named("paysmartID", "pay-4")).WillReturnRows(rows)

	repo := NewCardRepository(db)
	card, err := repo.FindByPaysmartID("pay-4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if card.PsProductCode != "" {
		t.Fatalf("expected empty ps_product_code when column is null, got %q", card.PsProductCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
