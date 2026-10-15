package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/DevThiagoSouza/src/application/services"
	"github.com/DevThiagoSouza/src/infra/db"
	httpapi "github.com/DevThiagoSouza/src/infra/http"
	"github.com/DevThiagoSouza/src/infra/logging"
	rabbitmqinfra "github.com/DevThiagoSouza/src/infra/messaging/rabbitmq"
	"github.com/DevThiagoSouza/src/infra/repository"
)

func main() {
	loadEnvFiles()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbConnection, err := db.ConnectionDB()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	sqlDB, err := dbConnection.DB()
	if err != nil {
		log.Fatalf("error accessing database connection: %v", err)
	}
	defer sqlDB.Close()

	appLogger := logging.NewSimpleLogger(logging.INFO)
	orderRepository := repository.NewOrderRepository(dbConnection, appLogger)
	orderService := services.NewOrderService(orderRepository)

	rabbitConn, rabbitCh, err := rabbitmqinfra.ConnectionRabbitMQ()
	if err != nil {
		log.Fatalf("error connecting to rabbitmq: %v", err)
	}
	defer rabbitCh.Close()
	defer rabbitConn.Close()

	queueName := getEnv("RABBITMQ_QUEUE", "orders")
	consumer := rabbitmqinfra.NewOrderConsumer(queueName, rabbitCh, orderService, appLogger)
	go func() {
		if err := consumer.Consume(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("rabbitmq consumer stopped: %v", err)
		}
	}()

	handler := httpapi.NewOrderHandler(orderService, orderService)
	router := httpapi.NewRouter(handler)
	server := &http.Server{
		Addr:    ":" + getEnv("HTTP_PORT", "8080"),
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("http server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadEnvFiles() {
	paths := []string{".env", filepath.Join("..", ".env")}
	for _, p := range paths {
		if err := loadEnvFile(p); err == nil {
			return
		}
	}
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")
		val = strings.Trim(val, "'")

		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
