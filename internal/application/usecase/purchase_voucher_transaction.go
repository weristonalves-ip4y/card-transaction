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
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("checking duplicate: %w", err)
	}
	if isDuplicate {
		output := rejectPurchaseByCode("07")
		_, _ = persistSerializedTransaction(a.txRepo, tx, output.Code)
		return output, nil
	}

	c, err := a.cardRepo.FindByPaysmartID(input.Card.PaysmartID)
	if err != nil {
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("loading card: %w", err)
	}

	if result := card.ValidateCard(c); !result.Approved {
		output := rejectPurchaseByCode(result.Code)
		_, _ = persistSerializedTransaction(a.txRepo, tx, output.Code)
		return output, nil
	}

	tx = tx.WithResolvedCard(c.ID, c.PsProductCode)
	tx = tx.WithResolvedAccountCard(c.AccountID, c.ID)

	if result := card.ValidateProductCompatibility(c, input.PsProductCode); !result.Approved {
		output := rejectPurchaseByCode(result.Code)
		_, _ = persistSerializedTransaction(a.txRepo, tx, output.Code)
		return output, nil
	}

	//TODO: O balance do voucher é outra tabela
	balance, err := a.balanceRepo.GetBalance(c.AccountID)
	if err != nil {
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("loading balance: %w", err)
	}

	if result := tx.ValidateBalance(balance); !result.Approved {
		output := rejectPurchaseByCode(result.Code)
		_, _ = persistSerializedTransaction(a.txRepo, tx, output.Code)
		return output, nil
	}

	amount := tx.Value
	remainingBalance, err := balance.Subtract(amount)
	if err != nil {
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("subtracting approved amount from voucher balance: %w", err)
	}

	approvedOutput := approvePurchase()
	authorizationID, err := persistSerializedTransaction(a.txRepo, tx, approvedOutput.Code)
	if err != nil {
		return rejectPurchaseByCode("96"), nil
	}

	balanceAmount := remainingBalance.Cents()
	approvedOutput.AuthorizationID = &authorizationID
	approvedOutput.BalanceAmount = &balanceAmount
	//TODO: Disparar o evento de debitar o valor daconta.
	//TODO: disparar evento de SMS para o cliente. Transação aprovado

	return approvedOutput, nil
}
