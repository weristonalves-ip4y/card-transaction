package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSystemStatusControllerHandleSuccess(t *testing.T) {
	t.Parallel()

	statusController := NewSystemStatusController(DBCheckFunc(func(ctx context.Context) error {
		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	statusController.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response SystemStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}

	if response.Code != 0 {
		t.Fatalf("expected code 0, got %d", response.Code)
	}

	if response.Message != "Operacao realizada com sucesso." {
		t.Fatalf("unexpected message: %s", response.Message)
	}
}

func TestSystemStatusControllerHandleDBError(t *testing.T) {
	t.Parallel()

	statusController := NewSystemStatusController(DBCheckFunc(func(ctx context.Context) error {
		return errors.New("db unavailable")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	statusController.Handle(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}

	var response SystemStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}

	if response.Code != 900 {
		t.Fatalf("expected code 900, got %d", response.Code)
	}

	if response.Message != "Sistema indisponivel. Erro ao acessar base de dados." {
		t.Fatalf("unexpected message: %s", response.Message)
	}
}

func TestSystemStatusControllerHandlePanic(t *testing.T) {
	t.Parallel()

	statusController := NewSystemStatusController(DBCheckFunc(func(ctx context.Context) error {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	statusController.Handle(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var response SystemStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}

	if response.Code != 999 {
		t.Fatalf("expected code 999, got %d", response.Code)
	}

	if response.Message != "Nao foi possivel executar comando. Erro desconhecido." {
		t.Fatalf("unexpected message: %s", response.Message)
	}
}

func TestSystemStatusControllerHandleMethodNotAllowed(t *testing.T) {
	t.Parallel()

	statusController := NewSystemStatusController(DBCheckFunc(func(ctx context.Context) error {
		return nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	statusController.Handle(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}

	if response.Error != "method not allowed" {
		t.Fatalf("unexpected error message: %s", response.Error)
	}
}
