package usecase

import (
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
	SaveSerialized(payload map[string]any) (int64, error)
}

type BalanceRepository interface {
	GetBalance(accountID int64) (vo.Money, error)
}

type BalanceVoucherRepository interface {
	GetBalanceVoucher(accountID int64) (vo.Money, error)
}

type CardMovementRepository interface {
	InsertDebitMovement(accountID, originID int64, movementTypeID int, amount float64, description string) error
}

type CardVoucherMovementRepository interface {
	InsertDebitVoucherMovement(accountID, originID int64, movementTypeID int, amount float64, description string, cardID, cardExternalID, cardPurchaseMcc interface{}) error
}

type PurchaseTransactionalContext interface {
	TransactionRepository() TransactionRepository
	CardMovementRepository() CardMovementRepository
	CardVoucherMovementRepository() CardVoucherMovementRepository
}

type PurchaseTransactionManager interface {
	WithinTransaction(fn func(ctx PurchaseTransactionalContext) error) error
}

type PurchaseOutput struct {
	Approved        bool
	Code            string
	Message         string
	Status          int
	AuthorizationID *int64
	BalanceAmount   *int64
}

type PurchaseTransaction struct {
	cardUseCase    PurchaseCardTransaction
	voucherUseCase PurchaseVoucherTransaction
}

func NewPurchaseTransaction(
	cardRepo CardRepository,
	txRepo TransactionRepository,
	balanceRepo BalanceRepository,
	balanceVoucherRepo BalanceVoucherRepository,
	movementRepo CardMovementRepository,
	movementVoucherRepo CardVoucherMovementRepository,
	txManager PurchaseTransactionManager,
) PurchaseTransaction {
	return PurchaseTransaction{
		cardUseCase:    NewPurchaseCardTransaction(cardRepo, txRepo, balanceRepo, movementRepo, txManager),
		voucherUseCase: NewPurchaseVoucherTransaction(cardRepo, txRepo, balanceVoucherRepo, movementVoucherRepo, txManager),
	}
}

func (a PurchaseTransaction) Execute(input dto.AuthorizePurchaseRequest) (PurchaseOutput, error) {
	if input.PsProductCode == "011401" {
		return a.voucherUseCase.Execute(input)
	} else {
		return a.cardUseCase.Execute(input)
	}

}
