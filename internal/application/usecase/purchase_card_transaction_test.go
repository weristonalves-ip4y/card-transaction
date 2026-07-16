package usecase

import (
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

func (r *txRepoSpy) ExistsByIdentifier(identifier string) (bool, error) {
	return r.existsByIdentifier, nil
}

func (r *txRepoSpy) GetMonthlySum(cardID int64) (vo.Money, error) {
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
	useCase := NewPurchaseCardTransaction(cardRepoStubForDuplicate{}, txRepo, balanceRepoStubForDuplicate{})

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

func validAuthorizePurchaseRequest() dto.AuthorizePurchaseRequest {
	creditCardAccount := "00"
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
			RequestMTI: "0100",
		},
		Establishment: dto.EstablishmentInput{},
		ForceAccept:   false,
		Internacional: false,
		Brand:         "visa",
		EntryMode:     "chip",
	}
}
