package purchase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"card-transaction/internal/application/dto"
	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/vo"
)

type txRepoSpy struct {
	existsByIdentifier bool
	saveCalls          int
	lastPayload        map[string]any
}

type movementRepoSpy struct {
	calls              int
	lastAccountID      int64
	lastOriginID       int64
	lastMovementTypeID int
	lastAmountCents    float64
	lastDescription    string
	err                error
}

type notificationDispatcherSpy struct {
	calls       int
	lastPhone   string
	lastMessage string
}

func (s *notificationDispatcherSpy) SendSMS(ctx context.Context, phone, message string) {
	s.calls++
	s.lastPhone = phone
	s.lastMessage = message
}

func (s *movementRepoSpy) InsertDebitMovement(accountID, originID int64, movementTypeID int, amount float64, description string) error {
	s.calls++
	s.lastAccountID = accountID
	s.lastOriginID = originID
	s.lastMovementTypeID = movementTypeID
	s.lastAmountCents = amount
	s.lastDescription = description
	return s.err
}

func (r *txRepoSpy) ExistsByIdentifier(identifier string) (bool, error) {
	return r.existsByIdentifier, nil
}

func (r *txRepoSpy) GetMonthlySum(cardID string) (vo.Money, error) {
	return vo.Zero(), nil
}

func (r *txRepoSpy) SaveSerialized(payload map[string]any) (int64, error) {
	r.saveCalls++
	r.lastPayload = payload
	return int64(r.saveCalls), nil
}

type cardRepoStubForDuplicate struct{}

func (s cardRepoStubForDuplicate) FindByPaysmartID(paysmartID string) (card.Card, error) {
	return card.Card{}, nil
}

type balanceRepoStubForDuplicate struct{}

func (s balanceRepoStubForDuplicate) GetBalance(accountID int64) (vo.Money, error) {
	return vo.Zero(), nil
}

func TestPurchaseCardTransactionExecuteDuplicatePersistsRejectedTransaction(t *testing.T) {
	t.Parallel()

	txRepo := &txRepoSpy{existsByIdentifier: true}
	useCase := NewPurchaseCardTransaction(cardRepoStubForDuplicate{}, txRepo, balanceRepoStubForDuplicate{}, &movementRepoSpy{})

	output, err := useCase.Execute(validAuthorizePurchaseRequest())
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

	if txRepo.lastPayload == nil {
		t.Fatalf("expected saved payload")
	}

	responseCode, ok := txRepo.lastPayload["response_code"].(string)
	if !ok {
		t.Fatalf("expected response_code field in payload")
	}

	if responseCode != "07" {
		t.Fatalf("expected persisted response_code 07, got %s", responseCode)
	}
}

type cardRepoApprovedStub struct{}

func (s cardRepoApprovedStub) FindByPaysmartID(paysmartID string) (card.Card, error) {
	return card.Card{
		ID:            11,
		AccountID:     22,
		CardID:        paysmartID,
		CardStatusID:  int(card.StatusActive),
		PsProductCode: "011202",
	}, nil
}

type balanceRepoApprovedStub struct{}

func (s balanceRepoApprovedStub) GetBalance(accountID int64) (vo.Money, error) {
	return vo.NewFromCents(5000)
}

func TestPurchaseCardTransactionExecuteApprovedCallsDebitMovement(t *testing.T) {
	t.Parallel()

	txRepo := &txRepoSpy{existsByIdentifier: false}
	movementRepo := &movementRepoSpy{}
	useCase := NewPurchaseCardTransaction(cardRepoApprovedStub{}, txRepo, balanceRepoApprovedStub{}, movementRepo)

	output, err := useCase.Execute(validAuthorizePurchaseRequest())
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

	responseCode, ok := txRepo.lastPayload["response_code"].(string)
	if !ok {
		t.Fatalf("expected response_code field in payload")
	}

	if responseCode != "00" {
		t.Fatalf("expected persisted response_code 00, got %s", responseCode)
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

	if movementRepo.lastMovementTypeID != cardPurchaseMovementTypeID {
		t.Fatalf("expected movement type %d, got %d", cardPurchaseMovementTypeID, movementRepo.lastMovementTypeID)
	}

	if movementRepo.lastAmountCents != 10 {
		t.Fatalf("expected movement amount 10.00, got %.2f", movementRepo.lastAmountCents)
	}

	if movementRepo.lastDescription != "COMPRA CARTÃO | MERCEARIA CENTRO" {
		t.Fatalf("expected movement description in upper case, got %s", movementRepo.lastDescription)
	}

}

func TestPurchaseCardTransactionExecuteMovementFailureKeepsApprovedPersistence(t *testing.T) {
	t.Parallel()

	txRepo := &txRepoSpy{existsByIdentifier: false}
	movementRepo := &movementRepoSpy{err: errors.New("db down")}
	useCase := NewPurchaseCardTransaction(cardRepoApprovedStub{}, txRepo, balanceRepoApprovedStub{}, movementRepo)

	_, err := useCase.Execute(validAuthorizePurchaseRequest())
	if err == nil {
		t.Fatalf("expected error when movement insert fails")
	}

	if txRepo.saveCalls != 1 {
		t.Fatalf("expected 1 persistence call, got %d", txRepo.saveCalls)
	}

	responseCode, ok := txRepo.lastPayload["response_code"].(string)
	if !ok {
		t.Fatalf("expected response_code field in payload")
	}

	if responseCode != "00" {
		t.Fatalf("expected persisted response_code 00, got %s", responseCode)
	}

	if movementRepo.calls != 1 {
		t.Fatalf("expected 1 movement call, got %d", movementRepo.calls)
	}
}

func TestPurchaseCardTransactionExecuteApprovedDispatchesSMSWhenPhoneProvided(t *testing.T) {
	t.Parallel()

	txRepo := &txRepoSpy{existsByIdentifier: false}
	movementRepo := &movementRepoSpy{}
	dispatcherSpy := &notificationDispatcherSpy{}

	useCase := NewPurchaseCardTransaction(cardRepoApprovedStub{}, txRepo, balanceRepoApprovedStub{}, movementRepo).
		WithNotificationDispatcher(dispatcherSpy)

	input := validAuthorizePurchaseRequest()
	phone := "5511999999999"
	input.Phone = &phone

	output, err := useCase.Execute(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !output.Approved {
		t.Fatalf("expected approved output")
	}

	if dispatcherSpy.calls != 1 {
		t.Fatalf("expected 1 sms dispatch call, got %d", dispatcherSpy.calls)
	}

	if dispatcherSpy.lastPhone != phone {
		t.Fatalf("expected sms phone %s, got %s", phone, dispatcherSpy.lastPhone)
	}

	if !strings.Contains(dispatcherSpy.lastMessage, "Compra CARTAO aprovada") {
		t.Fatalf("expected sms message to contain purchase approval content, got %s", dispatcherSpy.lastMessage)
	}
}

func validAuthorizePurchaseRequest() dto.AuthorizePurchaseRequest {
	creditCardAccount := "00"
	location := "Mercearia Centro"
	return dto.AuthorizePurchaseRequest{
		PurchaseID:        "tx-duplicate-1",
		AccountID:         "acc-1",
		PsProductCode:     "011402",
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
			RequestMTI:                      "0100",
			RequestCardAcceptorNameLocation: &location,
		},
		Establishment: dto.EstablishmentInput{},
		ForceAccept:   false,
		Internacional: false,
		Brand:         "visa",
		EntryMode:     "chip",
	}
}
