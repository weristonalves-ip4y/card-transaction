package router

import (
	"net/http"

	"card-transaction/internal/interfaces/http/controller"
)

func New(systemStatusController controller.SystemStatusController, authorizeController controller.AuthorizeController) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", systemStatusController.Handle)
	mux.HandleFunc("/", systemStatusController.Handle)

	mux.HandleFunc("/purchases", authorizeController.Handle)

	return mux
}
