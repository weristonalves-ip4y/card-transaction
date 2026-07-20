package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	var query string

	if strings.HasPrefix(paysmartID, "vcrt") {
		query = `
			SELECT TOP 1
				pc.id,
				pc.account_id,
				pc.uuid,
				pc.card_id,
				pc.card_status_id,
				pc.card_monthly_limit,
				pc.card_check_limit,
				pc.ps_product_code,
				pc.card_last_digit,
				pc.created_at,
				pc.updated_at,
				pc.deleted_at
			FROM
				paysmart_card_virtuals pcv
			LEFT JOIN paysmart_cards pc on
				pcv.paysmart_card_id = pc.id
			WHERE
			pcv.v_card_id = @paysmartID
			AND pcv.card_status_id = 4
			ORDER BY pc.id DESC
				`
		query = query // to avoid unused variable error
	} else {
		query = `
		SELECT TOP 1
			id,
			account_id,
			uuid,
			card_id,
			card_status_id,
			card_monthly_limit,
			card_check_limit,
			ps_product_code,
			card_last_digit,
			created_at,
			updated_at,
			deleted_at
		FROM paysmart_cards
		WHERE card_id = @paysmartID
		AND card_status_id = 4
		ORDER BY id DESC
	`
	}

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
		&c.FourLastDigits,
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

func (r CardRepository) FindOwnerPhoneByCardID(cardID int64) (string, error) {
	var phoneDDD, phoneNumber sql.NullString

	query := `
		SELECT TOP 1
		rc.owner_phone_ddd, rc.owner_phone_number 
		FROM request_cards rc 
		LEFT JOIN paysmart_cards pc 
		ON rc.id = pc.request_card_id 
		WHERE pc.id = @cardID
	`

	err := r.db.QueryRowContext(context.Background(), query, sql.Named("cardID", cardID)).Scan(&phoneDDD, &phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("querying owner phone by card id: %w", err)
	}

	if !phoneDDD.Valid || !phoneNumber.Valid {
		return "", nil
	}

	return "55" + phoneDDD.String + phoneNumber.String, nil
}
