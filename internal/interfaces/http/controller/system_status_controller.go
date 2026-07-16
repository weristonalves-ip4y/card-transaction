package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type DBChecker interface {
	Check(ctx context.Context) error
}

type SystemStatusController struct {
	checker DBChecker
}

func NewSystemStatusController(checker DBChecker) SystemStatusController {
	return SystemStatusController{checker: checker}
}

type SystemStatusResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (c SystemStatusController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("error no sistema: %v", recovered)
			writeJSON(w, http.StatusInternalServerError, SystemStatusResponse{
				Message: "Nao foi possivel executar comando. Erro desconhecido.",
				Code:    999,
			})
		}
	}()

	if c.checker == nil {
		log.Printf("Erro ao verificar status do sistema: db checker nao configurado")
		writeJSON(w, http.StatusServiceUnavailable, SystemStatusResponse{
			Message: "Sistema indisponivel. Erro ao acessar base de dados.",
			Code:    900,
		})
		return
	}

	if err := c.checker.Check(r.Context()); err != nil {
		log.Printf("Erro ao verificar status do sistema: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, SystemStatusResponse{
			Message: "Sistema indisponivel. Erro ao acessar base de dados.",
			Code:    900,
		})
		return
	}

	writeJSON(w, http.StatusOK, SystemStatusResponse{
		Message: "Operacao realizada com sucesso.",
		Code:    0,
	})
}

type DBCheckFunc func(ctx context.Context) error

func (f DBCheckFunc) Check(ctx context.Context) error {
	if f == nil {
		return fmt.Errorf("db checker function is nil")
	}

	return f(ctx)
}
