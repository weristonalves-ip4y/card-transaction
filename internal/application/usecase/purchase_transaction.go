package usecase

import (
	"fmt"

	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/vo"
)

type ValidationResult interface {
	Approved() bool
	Code() string
	Reason() string
}

type CardRepository interface {
	FindByPaysmartID(paysmartID string) (card.Card, error)
}

type TransactionRepository interface {
	ExistsByIdentifier(identifier string) (bool, error)
	GetMonthlySum(cardID string) (vo.Money, error)
	SaveSerialized(payload map[string]any) error
}

type BalanceRepository interface {
	GetBalance(accountID string) (vo.Money, error)
}

type PurchaseOutput struct {
	Approved bool
	Code     string
	Message  string
	Status   int
}

type PurchaseTransaction struct {
	cardUseCase    PurchaseCardTransaction
	voucherUseCase PurchaseVoucherTransaction
}

func NewPurchaseTransaction(
	cardRepo CardRepository,
	txRepo TransactionRepository,
	balanceRepo BalanceRepository,
) PurchaseTransaction {
	return PurchaseTransaction{
		cardUseCase:    NewPurchaseCardTransaction(cardRepo, txRepo, balanceRepo),
		voucherUseCase: NewPurchaseVoucherTransaction(cardRepo, txRepo, balanceRepo),
	}
}

func (a PurchaseTransaction) Execute(input dto.AuthorizePurchaseRequest) (PurchaseOutput, error) {
	if input.PsProductCode == "011401" {
		return a.voucherUseCase.Execute(input)
	} else if input.PsProductCode == "011402" {
		return a.cardUseCase.Execute(input)
	} else {
		return PurchaseOutput{}, fmt.Errorf("invalid product type: %s", input.PsProductCode)
	}

}
