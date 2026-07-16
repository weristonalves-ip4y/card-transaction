package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	cardMovementMasterID    = 1
	cardMovementTaxTypeID   = 25
	cardMovementTaxValueCts = int64(0)
)

type CardMovementRepository struct {
	db            *sql.DB
	procedureName string
}

func NewCardMovementRepository(db *sql.DB) (CardMovementRepository, error) {
	procedureName := envOrDefault("DB_CARD_MOVEMENT_PROCEDURE", "sp_insert_movement_card")
	if !isSafeIdentifier(procedureName) {
		return CardMovementRepository{}, fmt.Errorf("invalid DB_CARD_MOVEMENT_PROCEDURE: %q", procedureName)
	}

	return CardMovementRepository{db: db, procedureName: procedureName}, nil
}

func (r CardMovementRepository) InsertDebitMovement(accountID, originID int64, movementTypeID int, amount float64, description string) error {
	query := fmt.Sprintf(`EXEC [%s]
		@accountId = @accountId,
		@masterId = @masterId,
		@originId = @originId,
		@mvmntTypeId = @mvmntTypeId,
		@taxMvmntTypeId = @taxMvmntTypeId,
		@value = @value,
		@tax_value = @tax_value,
		@description = @description`, r.procedureName)

	_, err := r.db.ExecContext(
		context.Background(),
		query,
		sql.Named("accountId", accountID),
		sql.Named("masterId", cardMovementMasterID),
		sql.Named("originId", originID),
		sql.Named("mvmntTypeId", movementTypeID),
		sql.Named("taxMvmntTypeId", cardMovementTaxTypeID),
		sql.Named("value", amount),
		sql.Named("tax_value", cardMovementTaxValueCts),
		sql.Named("description", description),
	)
	if err != nil {
		return fmt.Errorf("inserting card movement for origin %d: %w", originID, err)
	}

	return nil
}

func (r CardMovementRepository) InsertCreditMovement(accountID, originID int64, movementTypeID int, amount float64, description string) error {
	amountNegative := -amount
	query := fmt.Sprintf(`EXEC [%s]
		@accountId = @accountId,
		@masterId = @masterId,
		@originId = @originId,
		@mvmntTypeId = @mvmntTypeId,
		@taxMvmntTypeId = @taxMvmntTypeId,
		@value = @value,
		@tax_value = @tax_value,
		@description = @description`, r.procedureName)

	_, err := r.db.ExecContext(
		context.Background(),
		query,
		sql.Named("accountId", accountID),
		sql.Named("masterId", cardMovementMasterID),
		sql.Named("originId", originID),
		sql.Named("mvmntTypeId", movementTypeID),
		sql.Named("taxMvmntTypeId", cardMovementTaxTypeID),
		sql.Named("value", amountNegative),
		sql.Named("tax_value", cardMovementTaxValueCts),
		sql.Named("description", description),
	)
	if err != nil {
		return fmt.Errorf("inserting card movement for origin %d: %w", originID, err)
	}

	return nil
}
