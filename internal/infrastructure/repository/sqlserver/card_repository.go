package sqlserver

import (
	"context"
	"database/sql"
	"fmt"

	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/vo"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return CardRepository{db: db}
}

func (r CardRepository) FindByPaysmartID(paysmartID string) (card.Card, error) {
	const query = `
		SELECT TOP 1
			id,
			account_id,
			uuid,
			card_id,
			card_status_id,
			card_monthly_limit,
			card_check_limit,
			ps_product_code,
			created_at,
			updated_at,
			deleted_at
		FROM paysmart_cards
		WHERE card_id = @paysmartID
		ORDER BY id DESC
	`

	var (
		c                 card.Card
		accountID         sql.NullInt64
		monthlyLimitCents sql.NullInt64
		checkLimit        sql.NullBool
		productCode       sql.NullString
	)

	err := r.db.QueryRowContext(context.Background(), query, sql.Named("paysmartID", paysmartID)).Scan(
		&c.ID,
		&accountID,
		&c.UUID,
		&c.CardID,
		&c.CardStatusID,
		&monthlyLimitCents,
		&checkLimit,
		&productCode,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return card.Card{}, nil
		}
		return card.Card{}, fmt.Errorf("querying card by paysmart id: %w", err)
	}

	resolvedMonthlyLimit := int64(0)
	if monthlyLimitCents.Valid {
		resolvedMonthlyLimit = monthlyLimitCents.Int64
	}

	if accountID.Valid {
		c.AccountID = accountID.Int64
	}

	monthlyLimit, err := vo.NewFromCents(resolvedMonthlyLimit)
	if err != nil {
		return card.Card{}, fmt.Errorf("invalid card_monthly_limit for card %d: %w", c.ID, err)
	}
	c.CardMonthlyLimit = monthlyLimit
	c.CardCheckLimit = checkLimit.Valid && checkLimit.Bool
	c.PsProductCode = productCode.String

	return c, nil
}
