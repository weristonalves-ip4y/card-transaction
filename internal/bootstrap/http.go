package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"card-transaction/internal/application/usecase"
	"card-transaction/internal/infrastructure/database"
	"card-transaction/internal/infrastructure/repository/memory"
	"card-transaction/internal/interfaces/http/controller"
	"card-transaction/internal/interfaces/http/router"
)

func NewHTTPHandler() http.Handler {
	cardRepo := memory.NewCardRepository()
	txRepo := memory.NewTransactionRepository()
	balanceRepo := memory.NewBalanceRepository()

	authorizeUseCase := usecase.NewPurchaseTransaction(cardRepo, txRepo, balanceRepo)
	authorizeController := controller.NewAuthorizeController(authorizeUseCase)
	systemStatusController := controller.NewSystemStatusController(controller.DBCheckFunc(func(ctx context.Context) error {
		cfg, err := database.LoadConfigFromEnv()
		if err != nil {
			return fmt.Errorf("loading db config: %w", err)
		}

		db, err := database.ConnectSQLServer(cfg)
		if err != nil {
			return fmt.Errorf("connecting to db: %w", err)
		}
		defer func() {
			_ = db.Close()
		}()

		if err := db.PingContext(ctx); err != nil {
			return fmt.Errorf("pinging db: %w", err)
		}

		return nil
	}))

	return router.New(systemStatusController, authorizeController)
}
