package usecase

import (
	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	domainnotification "card-transaction/internal/domain/notification"
	"fmt"
	"strings"
)

type PurchaseVoucherTransaction struct {
	cardRepo     CardRepository
	txRepo       TransactionRepository
	balanceRepo  BalanceVoucherRepository
	movementRepo CardVoucherMovementRepository
	txManager    TransactionManager
	dispatcher   domainnotification.Dispatcher
}

func NewPurchaseVoucherTransaction(
	cardRepo CardRepository,
	txRepo TransactionRepository,
	balanceRepo BalanceVoucherRepository,
	movementRepo CardVoucherMovementRepository,
	txManager ...TransactionManager,
) PurchaseVoucherTransaction {
	var resolvedTxManager TransactionManager
	if len(txManager) > 0 {
		resolvedTxManager = txManager[0]
	}

	return PurchaseVoucherTransaction{
		cardRepo:     cardRepo,
		txRepo:       txRepo,
		balanceRepo:  balanceRepo,
		movementRepo: movementRepo,
		txManager:    resolvedTxManager,
	}
}

func (a PurchaseVoucherTransaction) WithNotificationDispatcher(dispatcher domainnotification.Dispatcher) PurchaseVoucherTransaction {
	a.dispatcher = dispatcher
	return a
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
	var authorizationID int64
	persistenceCompleted := false
	movementDescription := buildCardVoucherPurchaseMovementDescription(input)

	if a.txManager != nil {
		err = a.txManager.WithinTransaction(func(ctx TransactionalContext) error {
			authID, txErr := persistSerializedTransaction(ctx.TransactionRepository(), tx, approvedOutput.Code)
			if txErr != nil {
				return txErr
			}
			persistenceCompleted = true

			if txErr := ctx.CardVoucherMovementRepository().InsertDebitVoucherMovement(
				c.AccountID,
				authID,
				cardPurchaseMovementTypeID,
				amount.ToFloat(),
				movementDescription,
				c.ID,
				c.CardID,
				input.OriginalIso8583.RequestMcc,
			); txErr != nil {
				return fmt.Errorf("inserting card voucher debit movement: %w", txErr)
			}

			authorizationID = authID
			return nil
		})
	} else {
		authorizationID, err = persistSerializedTransaction(a.txRepo, tx, approvedOutput.Code)
		if err == nil {
			persistenceCompleted = true
			err = a.movementRepo.InsertDebitVoucherMovement(
				c.AccountID,
				authorizationID,
				cardPurchaseMovementTypeID,
				amount.ToFloat(),
				movementDescription,
				c.ID,
				c.CardID,
				input.OriginalIso8583.RequestMcc,
			)
			if err != nil {
				err = fmt.Errorf("inserting card voucher debit movement: %w", err)
			}
		}
	}

	if err != nil {
		if !persistenceCompleted {
			return rejectPurchaseByCode("96"), nil
		}
		return PurchaseOutput{}, err
	}

	balanceAmount := remainingBalance.Cents()
	approvedOutput.AuthorizationID = &authorizationID
	approvedOutput.BalanceAmount = &balanceAmount
	notifyApprovedPurchaseSMS(a.dispatcher, "111", "voucher", amount, remainingBalance)

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
