package usecase

import (
	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	"fmt"
	"strings"
)

type PurchaseVoucherTransaction struct {
	cardRepo     CardRepository
	txRepo       TransactionRepository
	balanceRepo  BalanceVoucherRepository
	movementRepo CardVoucherMovementRepository
}

func NewPurchaseVoucherTransaction(
	cardRepo CardRepository,
	txRepo TransactionRepository,
	balanceRepo BalanceVoucherRepository,
	movementRepo CardVoucherMovementRepository,
) PurchaseVoucherTransaction {
	return PurchaseVoucherTransaction{
		cardRepo:     cardRepo,
		txRepo:       txRepo,
		balanceRepo:  balanceRepo,
		movementRepo: movementRepo,
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

	tx = tx.WithResolvedCard(c.CardID, c.PsProductCode)
	tx = tx.WithResolvedAccountCard(c.AccountID, c.CardID)

	if result := card.ValidateProductCompatibility(c, input.PsProductCode); !result.Approved {
		output := rejectPurchaseByCode(result.Code)
		_, _ = persistSerializedTransaction(a.txRepo, tx, output.Code)
		return output, nil
	}

	//TODO: O balance do voucher é outra tabela
	balance, err := a.balanceRepo.GetBalanceVoucher(c.AccountID)
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

	if err := a.movementRepo.InsertDebitVoucherMovement(
		c.AccountID,
		authorizationID,
		cardPurchaseMovementTypeID,
		amount.ToFloat(),
		buildCardVoucherPurchaseMovementDescription(input),
		c.ID,
		c.CardID,
		input.OriginalIso8583.RequestMcc,
	); err != nil {
		return PurchaseOutput{}, fmt.Errorf("inserting card voucher debit movement: %w", err)
	}
	//TODO: disparar evento de SMS para o cliente. Transação aprovado

	return approvedOutput, nil
}

func buildCardVoucherPurchaseMovementDescription(input dto.AuthorizePurchaseRequest) string {
	location := ""
	if input.OriginalIso8583.RequestCardAcceptorNameLocation != nil {
		location = strings.TrimSpace(*input.OriginalIso8583.RequestCardAcceptorNameLocation)
	}
	if location == "" && input.Establishment.Name != nil {
		location = strings.TrimSpace(*input.Establishment.Name)
	}

	if location == "" {
		return "COMPRA VOUCHER | ESTABELECIMENTO NÃO INFORMADO"
	}

	return "COMPRA VOUCHER | " + strings.ToUpper(location)
}
