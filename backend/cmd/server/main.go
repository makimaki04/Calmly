package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/makimaki04/Calmly/internal/config"
	"github.com/makimaki04/Calmly/internal/database"
	"github.com/makimaki04/Calmly/internal/handler"
	"github.com/makimaki04/Calmly/internal/logger"
	"github.com/makimaki04/Calmly/internal/middleware"
	"github.com/makimaki04/Calmly/internal/repository"
	"github.com/makimaki04/Calmly/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	appLogger, err := logger.InitLogger(cfg.LoggerConfig)
	if err != nil {
		log.Fatal("init logger: ", err)
	}
	defer func() {
		_ = appLogger.Sync()
	}()

	db, err := database.InitDB(context.Background(), cfg.DatabaseURI, cfg.DBConf, appLogger)
	if err != nil {
		// Error is logged inside repository layer (InitDB / migrations). Avoid duplicates here.
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewRepository(db, appLogger)
	svc := service.NewService(repo, cfg.DumpExpTime, appLogger)
	flow := service.NewFlowService(svc.Dump, svc.Analysis, svc.Answers, svc.Plan, svc.PlanItem, appLogger)
	httpHandler := handler.NewHandler(svc, flow, appLogger)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.WithLogging(appLogger))

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/sessions", httpHandler.StartSession)

		r.Route("/sessions/{dump_id}", func(r chi.Router) {
			r.Post("/answers", httpHandler.SubmitAnswers)
			r.Post("/plans/regenerate", httpHandler.RegeneratePlan)
			r.Post("/plans/finalize", httpHandler.FinalizePlanSelection)
		})
	})

	APIServer := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	go func() {
		if err := runServer(APIServer); err != nil {
			appLogger.Fatal(
				"http server startup error",
				zap.String("address", cfg.Address),
				zap.Error(err),
			)
		}
	}()

	gracefulStop(APIServer, appLogger, 5*time.Second)
}

func runServer(srv *http.Server) error {
	err := srv.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func gracefulStop(srv *http.Server, logger *zap.Logger, timeout time.Duration) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("received shutdown signal", zap.String("signal", ctx.Err().Error()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server graceful shutdown error", zap.Error(err))
		return
	}

	logger.Info("http server stopped gracefully")
}
