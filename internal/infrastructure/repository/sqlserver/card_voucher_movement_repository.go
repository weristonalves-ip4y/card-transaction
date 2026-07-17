package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	cardVoucherMovementMasterID    = 1
	cardVoucherMovementTaxTypeID   = 25
	cardVoucherMovementTaxValueCts = int64(0)
)

type CardVoucherMovementRepository struct {
	db            queryExecutor
	procedureName string
}

func NewCardVoucherMovementRepository(db *sql.DB) (CardVoucherMovementRepository, error) {
	procedureName := envOrDefault("DB_CARD_VOUCHER_MOVEMENT_PROCEDURE", "sp_insert_movement_voucher")
	if !isSafeIdentifier(procedureName) {
		return CardVoucherMovementRepository{}, fmt.Errorf("invalid DB_CARD_VOUCHER_MOVEMENT_PROCEDURE: %q", procedureName)
	}

	return newCardVoucherMovementRepositoryWithExecutor(db, procedureName), nil
}

func newCardVoucherMovementRepositoryWithExecutor(db queryExecutor, procedureName string) CardVoucherMovementRepository {
	return CardVoucherMovementRepository{db: db, procedureName: procedureName}
}

/**
* originId = transactionId
* accountId = cardId
* mvmntTypeId = movementTypeId
* value = amount
* description = description
* cardId = ID do cartão (interno do sistema)
* cardExternalId = cardId
* cardPurchaseMcc = requestMcc
**/

func (r CardVoucherMovementRepository) InsertDebitVoucherMovement(accountID, originID int64, movementTypeID int, amount float64, description string, cardID, cardExternalID, cardPurchaseMcc interface{}) error {
	query := fmt.Sprintf(`EXEC [%s]
	   	@masterId = @masterId,
		@originId = @originId,
		@accountId = @accountId,
		@mvmntTypeId = @mvmntTypeId,
		@value = @value,
		@description = @description,
		@cardId = @cardId,
		@cardExternalId = @cardExternalId,
		@cardPurchaseMcc = @cardPurchaseMcc`, r.procedureName)
	_, err := r.db.ExecContext(
		context.Background(),
		query,
		sql.Named("masterId", cardVoucherMovementMasterID),
		sql.Named("originId", originID),
		sql.Named("accountId", accountID),
		sql.Named("mvmntTypeId", movementTypeID),
		sql.Named("value", amount),
		sql.Named("description", description),
		sql.Named("cardId", cardID),
		sql.Named("cardExternalId", cardExternalID),
		sql.Named("cardPurchaseMcc", cardPurchaseMcc),
	)
	if err != nil {
		return fmt.Errorf("inserting card voucher movement for origin %d: %w", originID, err)
	}

	return nil
}

func (r CardVoucherMovementRepository) InsertCreditVoucherMovement(accountID, originID int64, movementTypeID int, amount float64, description string, cardID, cardExternalID, cardPurchaseMcc interface{}) error {
	amountNegative := -amount

	query := fmt.Sprintf(`EXEC [%s]
	   	@masterId = @masterId,
		@originId = @originId,
		@accountId = @accountId,
		@mvmntTypeId = @mvmntTypeId,
		@value = @value,
		@description = @description,
		@cardId = @cardId,
		@cardExternalId = @cardExternalId,
		@cardPurchaseMcc = @cardPurchaseMcc`, r.procedureName)
	_, err := r.db.ExecContext(
		context.Background(),
		query,
		sql.Named("masterId", cardVoucherMovementMasterID),
		sql.Named("originId", originID),
		sql.Named("accountId", accountID),
		sql.Named("mvmntTypeId", movementTypeID),
		sql.Named("value", amountNegative),
		sql.Named("description", description),
		sql.Named("cardId", cardID),
		sql.Named("cardExternalId", cardExternalID),
		sql.Named("cardPurchaseMcc", cardPurchaseMcc),
	)
	if err != nil {
		return fmt.Errorf("inserting card voucher movement for origin %d: %w", originID, err)
	}

	return nil
}
