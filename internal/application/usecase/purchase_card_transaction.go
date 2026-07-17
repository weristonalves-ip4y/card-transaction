package usecase

import (
	"fmt"
	"strings"

	"card-transaction/internal/application/decision"
	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/entity/transaction"
	"card-transaction/internal/domain/vo"

	"github.com/google/uuid"
)

const cardPurchaseMovementTypeID = 24

type PurchaseCardTransaction struct {
	cardRepo     CardRepository
	txRepo       TransactionRepository
	balanceRepo  BalanceRepository
	movementRepo CardMovementRepository
	txManager    TransactionManager
}

func NewPurchaseCardTransaction(
	cardRepo CardRepository,
	txRepo TransactionRepository,
	balanceRepo BalanceRepository,
	movementRepo CardMovementRepository,
	txManager ...TransactionManager,
) PurchaseCardTransaction {
	var resolvedTxManager TransactionManager
	if len(txManager) > 0 {
		resolvedTxManager = txManager[0]
	}

	return PurchaseCardTransaction{
		cardRepo:     cardRepo,
		txRepo:       txRepo,
		balanceRepo:  balanceRepo,
		movementRepo: movementRepo,
		txManager:    resolvedTxManager,
	}
}

func (a PurchaseCardTransaction) Execute(input dto.AuthorizePurchaseRequest) (PurchaseOutput, error) {
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

	monthlySum, err := a.txRepo.GetMonthlySum(c.CardID)
	if err != nil {
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("loading monthly sum: %w", err)
	}

	amount, err := vo.NewFromCents(input.TotalAmount.TotalAmount)
	if err != nil {
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("invalid transaction amount: %w", err)
	}

	if card.MonthsLimitExceeded(c, monthlySum, amount, input.ForceAccept) {
		output := rejectPurchaseByCode("01")
		_, _ = persistSerializedTransaction(a.txRepo, tx, output.Code)
		return output, nil
	}

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

	remainingBalance, err := balance.Subtract(amount)
	if err != nil {
		_, _ = persistSerializedTransaction(a.txRepo, tx, "96")
		return PurchaseOutput{}, fmt.Errorf("subtracting approved amount from balance: %w", err)
	}

	approvedOutput := approvePurchase()
	var authorizationID int64
	persistenceCompleted := false
	movementDescription := buildCardPurchaseMovementDescription(input)

	if a.txManager != nil {
		err = a.txManager.WithinTransaction(func(ctx TransactionalContext) error {
			authID, txErr := persistSerializedTransaction(ctx.TransactionRepository(), tx, approvedOutput.Code)
			if txErr != nil {
				return txErr
			}
			persistenceCompleted = true

			if txErr := ctx.CardMovementRepository().InsertDebitMovement(
				c.AccountID,
				authID,
				cardPurchaseMovementTypeID,
				amount.ToFloat(),
				movementDescription,
			); txErr != nil {
				return fmt.Errorf("inserting card debit movement: %w", txErr)
			}

			authorizationID = authID
			return nil
		})
	} else {
		authorizationID, err = persistSerializedTransaction(a.txRepo, tx, approvedOutput.Code)
		if err == nil {
			persistenceCompleted = true
			err = a.movementRepo.InsertDebitMovement(
				c.AccountID,
				authorizationID,
				cardPurchaseMovementTypeID,
				amount.ToFloat(),
				movementDescription,
			)
			if err != nil {
				err = fmt.Errorf("inserting card debit movement: %w", err)
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

	//TODO: disparar evento de SMS para o cliente. Transação aprovado

	return approvedOutput, nil
}

func buildCardPurchaseMovementDescription(input dto.AuthorizePurchaseRequest) string {
	location := ""
	if input.OriginalIso8583.RequestCardAcceptorNameLocation != nil {
		location = strings.TrimSpace(*input.OriginalIso8583.RequestCardAcceptorNameLocation)
	}
	if location == "" && input.Establishment.Name != nil {
		location = strings.TrimSpace(*input.Establishment.Name)
	}

	if location == "" {
		return "COMPRA CARTÃO"
	}

	return "COMPRA CARTÃO | " + strings.ToUpper(location)
}

func newPurchaseTransactionFromInput(input dto.AuthorizePurchaseRequest) (transaction.Transaction, error) {
	value, err := vo.NewFromCents(input.TotalAmount.TotalAmount)
	if err != nil {
		return transaction.Transaction{}, fmt.Errorf("invalid transaction value: %w", err)
	}

	tx, err := transaction.New(
		input.PurchaseID,
		transaction.TypePurchase,
		func() transaction.ProductType {
			if input.PsProductCode == "011401" {
				return transaction.ProductVoucher
			} else {
				return transaction.ProductCard
			}
		}(),
		value,
		input.ForceAccept,
	)
	if err != nil {
		return transaction.Transaction{}, fmt.Errorf("invalid transaction: %w", err)
	}

	originalValue, err := vo.NewFromCents(input.OriginalAmount.OriginalAmount)
	if err != nil {
		return transaction.Transaction{}, fmt.Errorf("invalid original amount: %w", err)
	}

	tx = tx.WithAuthorizeData(transaction.AuthorizeData{
		PurchaseID:                   input.PurchaseID,
		PsProductCode:                input.PsProductCode,
		PsProductName:                input.PsProductName,
		ProductType:                  input.ProductType,
		CountryCode:                  input.CountryCode,
		Source:                       input.Source,
		CallingSystemName:            input.CallingSystemName,
		Brand:                        input.Brand,
		EntryMode:                    input.EntryMode,
		HolderValidationMode:         input.HolderValidationMode,
		Internal:                     input.Internal,
		PreAuthorization:             input.PreAuthorization,
		IncrementalAuthorization:     input.IncrementalAuthorization,
		Nrid:                         input.Nrid,
		Authentication3DSTransaction: input.Authentication3DSTransaction,
		Platform:                     input.Platform,
		Internacional:                input.Internacional,
		AdditionalTerminalData: transaction.AdditionalTerminalData{
			TerminalType:                   input.AdditionalTerminalData.TerminalType,
			PartialApprovalIndicator:       input.AdditionalTerminalData.PartialApprovalIndicator,
			TerminalLocationIndicator:      input.AdditionalTerminalData.TerminalLocationIndicator,
			CardholderPresenceIndicator:    input.AdditionalTerminalData.CardholderPresenceIndicator,
			CardPresenceIndicator:          input.AdditionalTerminalData.CardPresenceIndicator,
			CardCaptureCapabilityIndicator: input.AdditionalTerminalData.CardCaptureCapabilityIndicator,
			TransactionStatusIndicator:     input.AdditionalTerminalData.TransactionStatusIndicator,
			TransactionSecurityIndicator:   input.AdditionalTerminalData.TransactionSecurityIndicator,
			TerminalPOSType:                input.AdditionalTerminalData.TerminalPOSType,
			TerminalInputCapability:        input.AdditionalTerminalData.TerminalInputCapability,
		},
		OriginalISO8583: transaction.ISORequest{
			MTI:   input.OriginalIso8583.RequestMTI,
			DE002: input.OriginalIso8583.RequestCardNumber,
			DE003: input.OriginalIso8583.RequestProcessingCode,
			DE004: input.OriginalIso8583.RequestTransactionAmountLocal,
			DE005: input.OriginalIso8583.RequestTransactionAmountReferencia,
			DE006: input.OriginalIso8583.RequestAmountInCardHolderBilling,
			DE007: input.OriginalIso8583.RequestTransmitionDateAndTime,
			DE008: input.OriginalIso8583.RequestAmountCardholderBillingFee,
			DE009: input.OriginalIso8583.RequestConversionRateSettlement,
			DE010: input.OriginalIso8583.RequestConvertionRateCardHolderBilling,
			DE011: input.OriginalIso8583.RequestSystemTraceAuditNumber,
			DE012: input.OriginalIso8583.RequestLocalTransactionTime,
			DE013: input.OriginalIso8583.RequestLocalTransactionDate,
			DE014: input.OriginalIso8583.RequestExpirationDate,
			DE015: input.OriginalIso8583.RequestSettlementDate,
			DE016: input.OriginalIso8583.RequestConversionDate,
			DE018: input.OriginalIso8583.RequestMcc,
			DE019: input.OriginalIso8583.RequestAcquiringInstitutionCountryCode,
			DE022: input.OriginalIso8583.RequestPosEntryMode,
			DE023: input.OriginalIso8583.RequestCardSequenceNumber,
			DE024: input.OriginalIso8583.RequestNetworkInternationalId,
			DE025: input.OriginalIso8583.RequestPosConditionCode,
			DE026: input.OriginalIso8583.RequestPosPinCaptureCode,
			DE028: input.OriginalIso8583.RequestTransactionFeeAmount,
			DE029: input.OriginalIso8583.RequestSettlementFeeAmount,
			DE032: input.OriginalIso8583.RequestAquiringInstitutionCode,
			DE033: input.OriginalIso8583.RequestForwardingInstitutionCode,
			DE035: input.OriginalIso8583.RequestTrack2Data,
			DE036: input.OriginalIso8583.RequestTrack3Data,
			DE037: input.OriginalIso8583.RequestRetrievalReferenceNumber,
			DE038: input.OriginalIso8583.RequestAuthorizationIdResponse,
			DE039: input.OriginalIso8583.RequestResponseCode,
			DE041: input.OriginalIso8583.RequestCardAcceptorTerminal,
			DE042: input.OriginalIso8583.RequestCardAcceptorIdentificationCode,
			DE043: input.OriginalIso8583.RequestCardAcceptorNameLocation,
			DE045: input.OriginalIso8583.RequestTrack1Data,
			DE046: input.OriginalIso8583.RequestAdditionalDataIso,
			DE047: input.OriginalIso8583.RequestAdditionalDataNational,
			DE048: input.OriginalIso8583.RequestContainsPdsInLtvFormat,
			DE049: input.OriginalIso8583.RequestTransactionCurrencyCode,
			DE050: input.OriginalIso8583.RequestTransactionAmount,
			DE051: input.OriginalIso8583.RequestCurrencyCodeCardholderBilling,
			DE052: input.OriginalIso8583.RequestPersonalIdentificationNumberData,
			DE053: input.OriginalIso8583.RequestSecurityRelatedControlInformation,
			DE054: input.OriginalIso8583.RequestAdditionalAmounts,
			DE055: input.OriginalIso8583.RequestIccSystemRelatedData,
			DE056: input.OriginalIso8583.RequestOriginalDataElements,
			DE058: input.OriginalIso8583.RequestAuthorizationLifeCycleCode,
			DE063: input.OriginalIso8583.RequestAuthorizingAgentIdentifierVoucher,
			DE122: input.OriginalIso8583.RequestAuthorizationAddressVoucher,
		},
		Card: transaction.CardData{
			PaysmartID: input.Card.PaysmartID,
			IssuerID:   input.Card.IssuerID,
			Pan:        input.Card.Pan,
			PanSeq:     input.Card.PanSeq,
			Bin:        input.Card.Bin,
		},
		Amount: transaction.AmountData{
			Total:                value,
			TotalCurrencyCode:    input.TotalAmount.CurrencyCode,
			Original:             originalValue,
			OriginalCurrencyCode: input.OriginalAmount.CurrencyCode,
		},
		ProcessingCode: transaction.ProcessingCode{
			TransactionType:        input.ProcessingCode.TipoTransacao,
			SourceAccountType:      input.ProcessingCode.SourceAccountType,
			CreditCardAccount:      input.ProcessingCode.CreditCardAccount,
			DestinationAccountType: input.ProcessingCode.DestinationAccountType,
		},
		Authorization: transaction.AuthorizationData{
			Code:        input.Authorization.Code,
			Description: input.Authorization.Description,
		},
		Establishment: transaction.EstablishmentData{
			MCC:     input.Establishment.Mcc,
			Name:    input.Establishment.Name,
			City:    input.Establishment.City,
			Address: input.Establishment.Address,
			State:   input.Establishment.State,
			Country: input.Establishment.Country,
			ZipCode: input.Establishment.ZipCode,
			Pat:     input.Establishment.Pat,
		},
	})

	tx = tx.WithPersistenceUUID(uuid.NewString())
	tx = tx.WithPersistenceBase(tx.UUID, input.Card.PaysmartID, &input.PurchaseID, nil)

	return tx, nil
}

func persistSerializedTransaction(repo TransactionRepository, tx transaction.Transaction, responseCode string) (int64, error) {
	tx = tx.BuildISOResponse(responseCode)
	return repo.SaveSerialized(tx.ToPersistenceMap())
}

func approvePurchase() PurchaseOutput {
	resolved := decision.Resolve("00")
	return PurchaseOutput{Approved: true, Code: "00", Message: resolved.Message, Status: resolved.HTTPStatus}
}

func rejectPurchaseByCode(code string) PurchaseOutput {
	resolved := decision.Resolve(code)
	return PurchaseOutput{Approved: false, Code: resolved.Code, Message: resolved.Message, Status: resolved.HTTPStatus}
}
