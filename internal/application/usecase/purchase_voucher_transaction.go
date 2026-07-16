package usecase

import (
	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	"fmt"
)

type PurchaseVoucherTransaction struct {
	cardRepo    CardRepository
	txRepo      TransactionRepository
	balanceRepo BalanceRepository
}

func NewPurchaseVoucherTransaction(
	cardRepo CardRepository,
	txRepo TransactionRepository,
	balanceRepo BalanceRepository,
) PurchaseVoucherTransaction {
	return PurchaseVoucherTransaction{
		cardRepo:    cardRepo,
		txRepo:      txRepo,
		balanceRepo: balanceRepo,
	}
}

func (a PurchaseVoucherTransaction) Execute(input dto.AuthorizePurchaseRequest) (PurchaseOutput, error) {
	tx, err := newPurchaseTransactionFromInput(input)
	if err != nil {
		return PurchaseOutput{}, err
	}

	isDuplicate, err := a.txRepo.ExistsByIdentifier(input.PurchaseID)
	if err != nil {
		return PurchaseOutput{}, fmt.Errorf("checking duplicate: %w", err)
	}
	if isDuplicate {
		_ = persistSerializedTransaction(a.txRepo, tx, "07")
		return rejectPurchaseByCode("07"), nil
	}

	c, err := a.cardRepo.FindByPaysmartID(input.Card.PaysmartID)
	if err != nil {
		return PurchaseOutput{}, fmt.Errorf("loading card: %w", err)
	}

	if result := card.ValidateCard(c); !result.Approved {
		_ = persistSerializedTransaction(a.txRepo, tx, result.Code)
		return rejectPurchaseByCode(result.Code), nil
	}

	tx = tx.WithResolvedCard(c.ID, c.ProductCode)
	tx = tx.WithResolvedAccountCard(c.AccountID, c.ID)

	if result := card.ValidateProductCompatibility(c, input.PsProductCode); !result.Approved {
		_ = persistSerializedTransaction(a.txRepo, tx, result.Code)
		return rejectPurchaseByCode(result.Code), nil
	}

	balance, err := a.balanceRepo.GetBalance(c.AccountID)
	if err != nil {
		return PurchaseOutput{}, fmt.Errorf("loading balance: %w", err)
	}

	if result := tx.ValidateBalance(balance); !result.Approved {
		_ = persistSerializedTransaction(a.txRepo, tx, result.Code)
		return rejectPurchaseByCode(result.Code), nil
	}

	_ = persistSerializedTransaction(a.txRepo, tx, "00")

	return approvePurchase(), nil
}
