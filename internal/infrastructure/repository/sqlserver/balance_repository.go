package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"card-transaction/internal/domain/vo"
)

var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type BalanceRepository struct {
	db           *sql.DB
	functionName string
	amountField  string
}

func NewBalanceRepository(db *sql.DB) (BalanceRepository, error) {
	functionName := envOrDefault("DB_BALANCE_FUNCTION", envOrDefault("DB_BALANCE_TABLE", "fn_get_account_balance"))
	amountField := envOrDefault("DB_BALANCE_AMOUNT_COLUMN", "balance")

	if !isSafeIdentifier(functionName) {
		return BalanceRepository{}, fmt.Errorf("invalid DB_BALANCE_FUNCTION: %q", functionName)
	}
	if !isSafeIdentifier(amountField) {
		return BalanceRepository{}, fmt.Errorf("invalid DB_BALANCE_AMOUNT_COLUMN: %q", amountField)
	}

	return BalanceRepository{
		db:           db,
		functionName: functionName,
		amountField:  amountField,
	}, nil
}

func (r BalanceRepository) GetBalance(accountID int64) (vo.Money, error) {
	query := fmt.Sprintf("SELECT TOP 1 %s FROM %s(@accountID)", r.amountField, r.functionName)

	var rawBalance any
	err := r.db.QueryRowContext(context.Background(), query, sql.Named("accountID", accountID)).Scan(&rawBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return vo.Zero(), nil
		}
		return vo.Money{}, fmt.Errorf("querying balance for account %d: %w", accountID, err)
	}

	cents, err := convertDBBalanceToCents(rawBalance)
	if err != nil {
		return vo.Money{}, fmt.Errorf("parsing balance for account %d: %w", accountID, err)
	}

	balance, err := vo.NewFromCents(cents)
	if err != nil {
		return vo.Money{}, fmt.Errorf("invalid balance value for account %d: %w", accountID, err)
	}

	return balance, nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func isSafeIdentifier(name string) bool {
	return identifierPattern.MatchString(name)
}

func convertDBBalanceToCents(raw any) (int64, error) {
	switch typed := raw.(type) {
	case nil:
		return 0, nil
	case int64:
		return typed, nil
	case int32:
		return int64(typed), nil
	case int:
		return int64(typed), nil
	case float64:
		return int64(math.Round(typed * 100)), nil
	case float32:
		return int64(math.Round(float64(typed) * 100)), nil
	case []byte:
		return parseBalanceStringToCents(string(typed))
	case string:
		return parseBalanceStringToCents(typed)
	default:
		return 0, fmt.Errorf("unsupported balance type %T", raw)
	}
}

func parseBalanceStringToCents(value string) (int64, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0, nil
	}

	if strings.Contains(normalized, ",") && !strings.Contains(normalized, ".") {
		normalized = strings.ReplaceAll(normalized, ",", ".")
	}

	if strings.Contains(normalized, ".") {
		amount, err := strconv.ParseFloat(normalized, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid decimal balance %q: %w", value, err)
		}
		return int64(math.Round(amount * 100)), nil
	}

	cents, err := strconv.ParseInt(normalized, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer balance %q: %w", value, err)
	}

	return cents, nil
}
