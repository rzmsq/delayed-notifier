package main

import (
	"context"
	"delayed-notifier/internal/appConfig"
	"delayed-notifier/internal/handler"
	"delayed-notifier/internal/rabbit"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/sync/errgroup"
)

const yamlPath = "config.yaml"

func main() {
	zlog.InitConsole()

	var configPath string
	flag.StringVar(&configPath, "config", yamlPath, "Path to appConfig file")
	flag.Parse()

	err := appConfig.InitConfigs(configPath)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Error loading appConfig file")
		os.Exit(1)
	}

	if err = run(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Error running server")
		os.Exit(1)
	}
}

func run() error {
	cfg := appConfig.Cfg

	conn, err := rabbitmq.Connect(cfg.RabbitConfig.RabbitmqUrl, cfg.RabbitConfig.RetryConnect, cfg.RabbitConfig.PauseRetry)
	if err != nil {
		return err
	}

	channelPool, err := rabbit.NewChannelPool(conn, 50)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	apiHandler := handler.APIHandler{Pool: channelPool}
	zlog.Logger.Info().Msgf("Starting API server %v", apiHandler.Pool)

	mux.HandleFunc("POST /notify", apiHandler.PostNotification)
	mux.HandleFunc("GET /notify/{id}", apiHandler.GetNotify)
	mux.HandleFunc("DELETE /notify/{id}", apiHandler.DeleteNotify)

	server := http.Server{
		Addr:         ":" + cfg.ServerConfig.Addr,
		Handler:      mux,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		<-groupCtx.Done()
		shCtx, cancel := context.WithTimeout(groupCtx, 60*time.Second)
		defer cancel()
		zlog.Logger.Info().Msgf("Shutting down graceful shutdown at %s", shCtx)
		if err := server.Shutdown(shCtx); err != nil {
			return fmt.Errorf("shutdown fail: %v", err)
		}
		return nil
	})

	group.Go(func() error {
		zlog.Logger.Info().Msgf("Starting server at %s", server.Addr)
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serrver faild: %v", err)
		}
		return nil
	})
	return group.Wait()
}
