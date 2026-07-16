package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"card-transaction/internal/application/dto"
	"card-transaction/internal/application/usecase"
)

type authorizerStub struct {
	executeFn func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error)
}

func (s authorizerStub) Execute(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
	return s.executeFn(input)
}

func TestAuthorizeControllerHandleHappyPath(t *testing.T) {
	t.Parallel()

	controller := NewAuthorizeController(authorizerStub{
		executeFn: func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
			if input.PurchaseID != "tx-1" {
				t.Fatalf("expected purchase_id tx-1, got %s", input.PurchaseID)
			}

			return usecase.PurchaseOutput{
				Approved: true,
				Code:     "00",
				Message:  "Operacao realizada com sucesso.",
				Status:   http.StatusOK,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewBuffer(validPurchaseRequestJSON(t, false)))
	rec := httptest.NewRecorder()

	controller.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if got := rec.Body.String(); got == "" {
		t.Fatalf("expected response body")
	}

	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid response json, got error: %v", err)
	}

	if response["authorization_code"] != "00" {
		t.Fatalf("expected authorization_code 00, got %v", response["authorization_code"])
	}

	if response["code"] != float64(0) {
		t.Fatalf("expected code 0, got %v", response["code"])
	}

	if _, ok := response["balance"]; !ok {
		t.Fatalf("expected balance in response")
	}

	if _, ok := response["authorization_id"]; !ok {
		t.Fatalf("expected authorization_id in response")
	}

	if _, ok := response["purchaseOnlyApproval"]; !ok {
		t.Fatalf("expected purchaseOnlyApproval in response")
	}
}

func TestAuthorizeControllerHandleUseCaseError(t *testing.T) {
	t.Parallel()

	controller := NewAuthorizeController(authorizerStub{
		executeFn: func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
			return usecase.PurchaseOutput{}, errors.New("db unavailable")
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewBuffer(validPurchaseRequestJSON(t, false)))
	rec := httptest.NewRecorder()

	controller.Handle(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid response json, got error: %v", err)
	}

	if response["authorization_code"] != "96" {
		t.Fatalf("expected authorization_code 96, got %v", response["authorization_code"])
	}

	if response["code"] != float64(900) {
		t.Fatalf("expected code 900, got %v", response["code"])
	}
}

func TestAuthorizeControllerHandleMethodNotAllowed(t *testing.T) {
	t.Parallel()

	controller := NewAuthorizeController(authorizerStub{
		executeFn: func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
			return usecase.PurchaseOutput{}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/purchases", nil)
	rec := httptest.NewRecorder()

	controller.Handle(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestAuthorizeControllerHandleInvalidPayload(t *testing.T) {
	t.Parallel()

	controller := NewAuthorizeController(authorizerStub{
		executeFn: func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
			t.Fatal("authorizer should not be called when payload is invalid")
			return usecase.PurchaseOutput{}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewBufferString(`{"purchase_id":`))
	rec := httptest.NewRecorder()

	controller.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthorizeControllerHandleUnknownFieldAllowed(t *testing.T) {
	t.Parallel()

	called := false
	controller := NewAuthorizeController(authorizerStub{
		executeFn: func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
			called = true
			return usecase.PurchaseOutput{
				Approved: true,
				Code:     "00",
				Message:  "Operacao realizada com sucesso.",
				Status:   http.StatusOK,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewBuffer(validPurchaseRequestJSON(t, true)))
	rec := httptest.NewRecorder()

	controller.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !called {
		t.Fatalf("expected authorizer to be called")
	}
}

func TestAuthorizeControllerHandleValidationError(t *testing.T) {
	t.Parallel()

	controller := NewAuthorizeController(authorizerStub{
		executeFn: func(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error) {
			t.Fatal("authorizer should not be called when validation fails")
			return usecase.PurchaseOutput{}, nil
		},
	})

	body := map[string]any{}
	body["purchase_id"] = "tx-1"

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewBuffer(encoded))
	rec := httptest.NewRecorder()

	controller.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func validPurchaseRequestJSON(t *testing.T, includeUnknownField bool) []byte {
	t.Helper()

	creditCardAccount := "00"
	body := dto.AuthorizePurchaseRequest{
		PurchaseID:        "tx-1",
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
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal valid request body: %v", err)
	}

	if includeUnknownField {
		var payload map[string]any
		if err := json.Unmarshal(encoded, &payload); err != nil {
			t.Fatalf("failed to unmarshal encoded request: %v", err)
		}

		payload["authenticationTokenizedMastercardData"] = map[string]any{"eci": "05"}

		encodedWithUnknownField, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal request with unknown field: %v", err)
		}

		return encodedWithUnknownField
	}

	return encoded
}
