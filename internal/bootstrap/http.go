package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"card-transaction/internal/application/usecase"
	"card-transaction/internal/infrastructure/database"
	"card-transaction/internal/infrastructure/repository/sqlserver"
	"card-transaction/internal/interfaces/http/controller"
	"card-transaction/internal/interfaces/http/router"
)

func NewHTTPHandler() http.Handler {
	cfg, err := database.LoadConfigFromEnv()
	if err != nil {
		log.Panicf("loading db config: %v", err)
	}

	db, err := database.ConnectSQLServer(cfg)
	if err != nil {
		log.Panicf("connecting to db: %v", err)
	}

	cardRepo := sqlserver.NewCardRepository(db)
	txRepo := sqlserver.NewTransactionRepository(db)
	balanceRepo, err := sqlserver.NewBalanceRepository(db)
	if err != nil {
		log.Panicf("building sql balance repository: %v", err)
	}

	balanceVoucherRepo, err := sqlserver.NewBalanceVoucherRepository(db)
	if err != nil {
		log.Panicf("building sql balance voucher repository: %v", err)
	}
	movementRepo, err := sqlserver.NewCardMovementRepository(db)
	if err != nil {
		log.Panicf("building sql card movement repository: %v", err)
	}
	movementVoucherRepo, err := sqlserver.NewCardVoucherMovementRepository(db)
	if err != nil {
		log.Panicf("building sql card voucher movement repository: %v", err)
	}
	transactionManager, err := sqlserver.NewSQLServerTransactionManager(db)
	if err != nil {
		log.Panicf("building sql purchase tx manager: %v", err)
	}

	authorizeUseCase := usecase.NewPurchaseTransaction(cardRepo, txRepo, balanceRepo, balanceVoucherRepo, movementRepo, movementVoucherRepo, transactionManager)
	authorizeController := controller.NewAuthorizeController(authorizeUseCase)
	systemStatusController := controller.NewSystemStatusController(controller.DBCheckFunc(func(ctx context.Context) error {
		if err := db.PingContext(ctx); err != nil {
			return fmt.Errorf("pinging db: %w", err)
		}

		return nil
	}))

	return router.New(systemStatusController, authorizeController)
}
