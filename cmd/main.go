package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"card-transaction/internal/bootstrap"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	readHeaderTimeout := getEnvSeconds("HTTP_READ_HEADER_TIMEOUT", 5)
	readTimeout := getEnvSeconds("HTTP_READ_TIMEOUT", 10)
	writeTimeout := getEnvSeconds("HTTP_WRITE_TIMEOUT", 15)
	idleTimeout := getEnvSeconds("HTTP_IDLE_TIMEOUT", 60)
	shutdownTimeout := getEnvSeconds("HTTP_SHUTDOWN_TIMEOUT", 10)

	handler := bootstrap.NewHTTPHandler()

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		log.Printf("API listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownCh

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = server.Close()
	}

	log.Println("server stopped")
}

func getEnvSeconds(key string, fallback int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallback) * time.Second
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		log.Printf("invalid value for %s=%q, using %ds", key, value, fallback)
		return time.Duration(fallback) * time.Second
	}

	return time.Duration(seconds) * time.Second
}
