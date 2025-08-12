package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	//_ "github.com/senyabanana/pvz-service/docs"
	"github.com/senyabanana/pvz-service/internal/handler"
	"github.com/senyabanana/pvz-service/internal/infrastructure/config"
	"github.com/senyabanana/pvz-service/internal/infrastructure/database"
	"github.com/senyabanana/pvz-service/internal/infrastructure/logger"
	"github.com/senyabanana/pvz-service/internal/repository"
	"github.com/senyabanana/pvz-service/internal/service"
	grpcServer "github.com/senyabanana/pvz-service/internal/transport/grpc"
	httpServer "github.com/senyabanana/pvz-service/internal/transport/http"
)

// @title PVZ Service API
// @version 1.0
// @description API для управления пунктами выдачи заказов (ПВЗ), приёмками и товарами.
// @BasePath /
// @schemes http
// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	log := logger.NewLogger()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}

	db, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}

	defer db.Close()

	trManager := manager.Must(trmsqlx.NewDefaultFactory(db))
	repos := repository.NewRepository(db)
	services := service.NewService(repos, trManager, cfg.JWTSecretKey, log)
	handlers := handler.NewHandler(services, cfg.JWTSecretKey, log)
	routes := httpServer.SetupRouter(handlers, cfg.JWTSecretKey, log)
	httpSrv := httpServer.NewServer(routes, cfg.ServerPort, log)
	grpcSrv, err := grpcServer.NewGRPCServer(cfg.GRPCPort, services, log)
	if err != nil {
		log.Fatalf("grpc server init failed: %v", err)
	}

	go func() {
		if err := httpSrv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		if err := grpcSrv.Run(); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server shutdown error: %s", err.Error())
	}

	grpcSrv.Shutdown(shutdownCtx)

	log.Info("Service stopped gracefully")
}
