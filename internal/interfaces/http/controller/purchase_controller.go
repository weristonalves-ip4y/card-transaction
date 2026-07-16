package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"card-transaction/internal/application/decision"
	"card-transaction/internal/application/dto"
	"card-transaction/internal/application/usecase"
)

/**
* Controller responsavel por receber a requisição de autorização de compra, validar os dados e chamar o caso de uso correspondente.
* Tipo: Purchase (Compra e Voucher)
**/

type Authorizer interface {
	Execute(input dto.AuthorizePurchaseRequest) (usecase.PurchaseOutput, error)
}

type AuthorizeController struct {
	authorizer Authorizer
}

func NewAuthorizeController(authorizer Authorizer) AuthorizeController {
	return AuthorizeController{authorizer: authorizer}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (c AuthorizeController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		log.Printf("purchase request read body error: method=%s path=%s err=%v", r.Method, r.URL.Path, err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request payload",
		})
		return
	}

	_ = r.Body.Close()

	log.Printf("purchase request received: method=%s path=%s body=%s", r.Method, r.URL.Path, sanitizePayloadForLog(body))

	var req dto.AuthorizePurchaseRequest

	decoder := json.NewDecoder(bytes.NewReader(body))

	if err := decoder.Decode(&req); err != nil {
		log.Printf("purchase request decode error: method=%s path=%s err=%v", r.Method, r.URL.Path, err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request payload",
		})
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("purchase request validation error: method=%s path=%s purchase_id=%s account_id=%s err=%v", r.Method, r.URL.Path, req.PurchaseID, req.AccountID, err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	output, err := c.authorizer.Execute(req)
	if err != nil {
		log.Printf("purchase request usecase error: method=%s path=%s purchase_id=%s account_id=%s err=%v", r.Method, r.URL.Path, req.PurchaseID, req.AccountID, err)
		resolved := decision.Resolve("96")
		response := dto.PurchaseOutputFromAuthorizationCode(resolved.Code, resolved.Message)
		writeJSON(w, response.StatusCode, response.Data)
		return
	}

	log.Printf("purchase request completed: method=%s path=%s purchase_id=%s account_id=%s approved=%t code=%s status=%d", r.Method, r.URL.Path, req.PurchaseID, req.AccountID, output.Approved, output.Code, output.Status)

	response := dto.PurchaseOutputFromAuthorizationCode(output.Code, output.Message)
	writeJSON(w, response.StatusCode, response.Data)
}

func sanitizePayloadForLog(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return string(body)
	}

	cardValue, ok := payload["card"]
	if ok {
		cardMap, ok := cardValue.(map[string]any)
		if ok {
			panValue, ok := cardMap["pan"]
			if ok {
				pan, ok := panValue.(string)
				if ok {
					cardMap["pan"] = maskPan(pan)
				}
			}
		}
	}

	sanitized, err := json.Marshal(payload)
	if err != nil {
		return string(body)
	}

	return string(sanitized)
}

func maskPan(pan string) string {
	pan = strings.TrimSpace(pan)
	if len(pan) <= 10 {
		return "***"
	}

	return pan[:6] + strings.Repeat("*", len(pan)-10) + pan[len(pan)-4:]
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
