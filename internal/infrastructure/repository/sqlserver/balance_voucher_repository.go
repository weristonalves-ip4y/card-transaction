package sqlserver

import (
	"context"
	"database/sql"
	"fmt"

	"card-transaction/internal/domain/vo"
)

type BalanceVoucherRepository struct {
	db           *sql.DB
	functionName string
	amountField  string
}

func NewBalanceVoucherRepository(db *sql.DB) (BalanceVoucherRepository, error) {
	functionName := envOrDefault(
		"DB_VOUCHER_BALANCE_FUNCTION",
		envOrDefault("DB_BALANCE_FUNCTION", envOrDefault("DB_VOUCHER_BALANCE_TABLE", envOrDefault("DB_BALANCE_TABLE", "fn_get_voucher_balance"))),
	)
	amountField := envOrDefault("DB_VOUCHER_BALANCE_AMOUNT_COLUMN", envOrDefault("DB_BALANCE_AMOUNT_COLUMN", "balance"))

	if !isSafeIdentifier(functionName) {
		return BalanceVoucherRepository{}, fmt.Errorf("invalid DB_VOUCHER_BALANCE_FUNCTION: %q", functionName)
	}
	if !isSafeIdentifier(amountField) {
		return BalanceVoucherRepository{}, fmt.Errorf("invalid DB_VOUCHER_BALANCE_AMOUNT_COLUMN: %q", amountField)
	}

	return BalanceVoucherRepository{
		db:           db,
		functionName: functionName,
		amountField:  amountField,
	}, nil
}

func (r BalanceVoucherRepository) GetBalanceVoucher(accountID int64) (vo.Money, error) {
	query := fmt.Sprintf("SELECT TOP 1 %s FROM %s(@accountID)", r.amountField, r.functionName)

	var rawBalance any
	err := r.db.QueryRowContext(context.Background(), query, sql.Named("accountID", accountID)).Scan(&rawBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return vo.Zero(), nil
		}
		return vo.Money{}, fmt.Errorf("querying balance voucher for account %d: %w", accountID, err)
	}

	cents, err := convertDBBalanceToCents(rawBalance)
	if err != nil {
		return vo.Money{}, fmt.Errorf("parsing balance voucher for account %d: %w", accountID, err)
	}

	balance, err := vo.NewFromCents(cents)
	if err != nil {
		return vo.Money{}, fmt.Errorf("invalid balance voucher value for account %d: %w", accountID, err)
	}

	return balance, nil
}
