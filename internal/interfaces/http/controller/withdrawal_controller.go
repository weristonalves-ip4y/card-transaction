package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"card-transaction/internal/application/decision"
	"card-transaction/internal/application/dto"
)

/**
* Controller responsavel por receber a requisição de autorização de saque, validar os dados e chamar o caso de uso correspondente.
* Tipo: Withdrawal (Saque)
**/

type WithdrawalController struct {
	authorizer Authorizer
}

func NewWithdrawalController(authorizer Authorizer) WithdrawalController {
	return WithdrawalController{authorizer: authorizer}
}

func (c WithdrawalController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		log.Printf("withdrawal request read body error: method=%s path=%s err=%v", r.Method, r.URL.Path, err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request payload",
		})
		return
	}

	_ = r.Body.Close()

	log.Printf("withdrawal request received: method=%s path=%s body=%s", r.Method, r.URL.Path, sanitizePayloadForLog(body))

	var req dto.AuthorizeRequest

	decoder := json.NewDecoder(bytes.NewReader(body))

	if err := decoder.Decode(&req); err != nil {
		log.Printf("withdrawal request decode error: method=%s path=%s err=%v", r.Method, r.URL.Path, err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request payload",
		})
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("withdrawal request validation error: method=%s path=%s withdrawal_id=%s account_id=%s err=%v", r.Method, r.URL.Path, *req.WithdrawalID, req.AccountID, err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if req.Authorization.Code != "00" {
		log.Printf("withdrawal request denied before usecase: method=%s path=%s withdrawal_id=%s account_id=%s authorization_code=%s", r.Method, r.URL.Path, *req.WithdrawalID, req.AccountID, req.Authorization.Code)
		response := dto.TransactionOutputFromIncomingDenial(req.Authorization.Code, req.Authorization.Description)
		writeJSON(w, response.StatusCode, response.Data)
		return
	}

	output, err := c.authorizer.Execute(req)
	if err != nil {
		log.Printf("withdrawal request usecase error: method=%s path=%s withdrawal_id=%s account_id=%s err=%v", r.Method, r.URL.Path, req.WithdrawalID, req.AccountID, err)
		resolved := decision.Resolve("96")
		response := dto.TransactionOutputFromAuthorizationCode(resolved.Code, resolved.Message)
		writeJSON(w, response.StatusCode, response.Data)
		return
	}

	log.Printf("withdrawal request completed: method=%s path=%s withdrawal_id=%s account_id=%s approved=%t code=%s status=%d", r.Method, r.URL.Path, *req.WithdrawalID, req.AccountID, output.Approved, output.Code, output.Status)

	response := dto.TransactionOutputFromAuthorizationCode(output.Code, output.Message)
	if output.AuthorizationID != nil {
		response.Data["authorization_id"] = *output.AuthorizationID
	}

	if output.BalanceAmount != nil {
		response.Data["balance"] = map[string]any{
			"amount":        *output.BalanceAmount,
			"currency_code": "986",
		}
	}

	writeJSON(w, response.StatusCode, response.Data)
}
