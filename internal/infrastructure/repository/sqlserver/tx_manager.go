package sqlserver

import (
	"context"
	"database/sql"
	"fmt"

	"card-transaction/internal/application/usecase"
)

type SQLServerTransactionManager struct {
	db                       *sql.DB
	cardMovementProcedure    string
	voucherMovementProcedure string
}

type purchaseTxContext struct {
	txRepo                  usecase.TransactionRepository
	cardMovementRepo        usecase.CardMovementRepository
	cardVoucherMovementRepo usecase.CardVoucherMovementRepository
}

func (c purchaseTxContext) TransactionRepository() usecase.TransactionRepository {
	return c.txRepo
}

func (c purchaseTxContext) CardMovementRepository() usecase.CardMovementRepository {
	return c.cardMovementRepo
}

func (c purchaseTxContext) CardVoucherMovementRepository() usecase.CardVoucherMovementRepository {
	return c.cardVoucherMovementRepo
}

func NewSQLServerTransactionManager(db *sql.DB) (SQLServerTransactionManager, error) {
	cardMovementRepo, err := NewCardMovementRepository(db)
	if err != nil {
		return SQLServerTransactionManager{}, fmt.Errorf("building card movement repository for tx manager: %w", err)
	}

	voucherMovementRepo, err := NewCardVoucherMovementRepository(db)
	if err != nil {
		return SQLServerTransactionManager{}, fmt.Errorf("building card voucher movement repository for tx manager: %w", err)
	}

	return SQLServerTransactionManager{
		db:                       db,
		cardMovementProcedure:    cardMovementRepo.procedureName,
		voucherMovementProcedure: voucherMovementRepo.procedureName,
	}, nil
}

func (m SQLServerTransactionManager) WithinTransaction(fn func(ctx usecase.TransactionalContext) error) error {
	tx, err := m.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	ctx := purchaseTxContext{
		txRepo:                  newTransactionRepositoryWithExecutor(tx),
		cardMovementRepo:        newCardMovementRepositoryWithExecutor(tx, m.cardMovementProcedure),
		cardVoucherMovementRepo: newCardVoucherMovementRepositoryWithExecutor(tx, m.voucherMovementProcedure),
	}

	if err := fn(ctx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rolling back transaction after error %v: %w", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("commit failed (%v) and rollback failed: %w", err, rollbackErr)
		}
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
