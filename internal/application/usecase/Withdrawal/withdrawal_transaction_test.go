package withdrawal

import (
	"testing"

	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/vo"
)

type withdrawalTxRepoSpy struct {
	existsByIdentifier bool
	saveCalls          int
	lastPayload        map[string]any
}

func (r *withdrawalTxRepoSpy) ExistsByIdentifier(identifier *string) (bool, error) {
	if identifier == nil {
		return false, nil
	}
	return r.existsByIdentifier, nil
}

func (r *withdrawalTxRepoSpy) GetMonthlySum(cardID string) (vo.Money, error) {
	return vo.Zero(), nil
}

func (r *withdrawalTxRepoSpy) SaveSerialized(payload map[string]any) (int64, error) {
	r.saveCalls++
	r.lastPayload = payload
	return int64(r.saveCalls), nil
}

type withdrawalCardRepoStub struct{}

func (s withdrawalCardRepoStub) FindByPaysmartID(paysmartID string) (card.Card, error) {
	return card.Card{
		ID:            11,
		AccountID:     22,
		CardID:        paysmartID,
		CardStatusID:  int(card.StatusActive),
		PsProductCode: "011202",
	}, nil
}

func (s withdrawalCardRepoStub) FindOwnerPhoneByCardID(cardID int64) (string, error) {
	return "", nil
}

type withdrawalBalanceRepoStub struct{}

func (s withdrawalBalanceRepoStub) GetBalance(accountID int64) (vo.Money, error) {
	return vo.NewFromCents(5000)
}

type withdrawalMovementRepoSpy struct {
	calls              int
	lastAccountID      int64
	lastOriginID       int64
	lastMovementTypeID int
	lastAmount         float64
	lastDescription    string
}

func (s *withdrawalMovementRepoSpy) InsertDebitMovement(accountID, originID int64, movementTypeID int, amount float64, description string) error {
	s.calls++
	s.lastAccountID = accountID
	s.lastOriginID = originID
	s.lastMovementTypeID = movementTypeID
	s.lastAmount = amount
	s.lastDescription = description
	return nil
}

func TestWithdrawalCardTransactionExecuteApprovedPersistsWithdrawalIDAndMovement(t *testing.T) {
	t.Parallel()

	txRepo := &withdrawalTxRepoSpy{existsByIdentifier: false}
	movementRepo := &withdrawalMovementRepoSpy{}
	useCase := NewWithdrawalTransaction(withdrawalCardRepoStub{}, txRepo, withdrawalBalanceRepoStub{}, movementRepo, nil)

	input := validAuthorizeWithdrawalRequest()

	output, err := useCase.Execute(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !output.Approved {
		t.Fatalf("expected approved output")
	}

	if output.Code != "00" {
		t.Fatalf("expected code 00, got %s", output.Code)
	}

	if txRepo.saveCalls != 1 {
		t.Fatalf("expected 1 persistence call, got %d", txRepo.saveCalls)
	}

	withdrawalIDPtr, ok := txRepo.lastPayload["withdrawal_id"].(*string)
	if !ok || withdrawalIDPtr == nil {
		t.Fatalf("expected persisted withdrawal_id pointer")
	}

	if *withdrawalIDPtr != "wd-1" {
		t.Fatalf("expected persisted withdrawal_id wd-1, got %s", *withdrawalIDPtr)
	}

	if movementRepo.calls != 1 {
		t.Fatalf("expected 1 movement call, got %d", movementRepo.calls)
	}

	if movementRepo.lastAccountID != 22 {
		t.Fatalf("expected movement account 22, got %d", movementRepo.lastAccountID)
	}

	if movementRepo.lastOriginID != 1 {
		t.Fatalf("expected movement origin 1, got %d", movementRepo.lastOriginID)
	}

	if movementRepo.lastMovementTypeID != cardWithdrawalMovementTypeID {
		t.Fatalf("expected movement type %d, got %d", cardWithdrawalMovementTypeID, movementRepo.lastMovementTypeID)
	}

	if movementRepo.lastAmount != 10 {
		t.Fatalf("expected movement amount 10.00, got %.2f", movementRepo.lastAmount)
	}
}

func TestWithdrawalCardTransactionExecuteDuplicatePersistsRejectedTransaction(t *testing.T) {
	t.Parallel()

	txRepo := &withdrawalTxRepoSpy{existsByIdentifier: true}
	useCase := NewWithdrawalTransaction(withdrawalCardRepoStub{}, txRepo, withdrawalBalanceRepoStub{}, &withdrawalMovementRepoSpy{}, nil)

	output, err := useCase.Execute(validAuthorizeWithdrawalRequest())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if output.Approved {
		t.Fatalf("expected rejected output")
	}

	if output.Code != "07" {
		t.Fatalf("expected code 07, got %s", output.Code)
	}

	if txRepo.saveCalls != 1 {
		t.Fatalf("expected 1 persistence call, got %d", txRepo.saveCalls)
	}

	responseCode, ok := txRepo.lastPayload["response_code"].(string)
	if !ok {
		t.Fatalf("expected response_code field in payload")
	}

	if responseCode != "07" {
		t.Fatalf("expected persisted response_code 07, got %s", responseCode)
	}
}

func validAuthorizeWithdrawalRequest() dto.AuthorizeRequest {
	creditCardAccount := "00"
	purchaseID := "purchase-1"
	withdrawalID := "wd-1"
	location := "ATM Centro"

	return dto.AuthorizeRequest{
		PurchaseID:        &purchaseID,
		WithdrawalID:      &withdrawalID,
		AccountID:         "acc-1",
		PsProductCode:     "011202",
		PsProductName:     "card",
		ProductType:       "card",
		CountryCode:       "076",
		Source:            "api",
		CallingSystemName: "payments",
		Card: dto.CardInput{
			PaysmartID: "pay-1",
			IssuerID:   "iss-1",
			Pan:        "4111111111111111",
			PanSeq:     "1",
			Bin:        "411111",
		},
		TotalAmount: dto.TotalAmountInput{
			TotalAmount:  1000,
			CurrencyCode: 986,
		},
		OriginalAmount: dto.OriginalAmountInput{
			OriginalAmount: 1000,
			CurrencyCode:   986,
		},
		ProcessingCode: dto.ProcessingCodeInput{
			TipoTransacao:          "00",
			SourceAccountType:      "00",
			CreditCardAccount:      &creditCardAccount,
			DestinationAccountType: "00",
		},
		Authorization: dto.Authorization{
			Code:        "00",
			Description: "approved",
		},
		Fees: []dto.FeesInput{},
		OriginalIso8583: dto.OriginalIso8583{
			RequestMTI:                      "0200",
			RequestCardAcceptorNameLocation: &location,
		},
		Establishment: dto.EstablishmentInput{},
		ForceAccept:   false,
		Internacional: false,
		Brand:         "visa",
		EntryMode:     "chip",
	}
}
