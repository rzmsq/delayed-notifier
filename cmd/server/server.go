package main

import (
	"context"
	"delayed-notifier/internal/handler/notifier"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/sync/errgroup"
)

const yamlPath = "config.yaml"

type Config struct {
	Port string `yaml:"port" env:"DELAYED_NOTIFIER_PORT" envDefault:"8080"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", yamlPath, "Path to config file")
	flag.Parse()

	var config Config
	err := cleanenv.ReadConfig(configPath, &config)
	if err != nil {
		panic(err)
	}

	if err = run(config); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(config Config) error {
	mux := http.NewServeMux()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux.HandleFunc("POST /notify", notifier.CreateNotify)
	mux.HandleFunc("GET /notify/{id}", notifier.GetNotify)
	mux.HandleFunc("DELETE /notify/{id}", notifier.DeleteNotify)

	server := http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		<-groupCtx.Done()
		shCtx, cancel := context.WithTimeout(groupCtx, 60*time.Second)
		defer cancel()
		slog.Info("Shutting down server...")
		if err := server.Shutdown(shCtx); err != nil {
			return fmt.Errorf("shutdown fail: %v", err)
		}
		return nil
	})

	group.Go(func() error {
		slog.Info("Server starting on port " + server.Addr)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serrver faild: %v", err)
		}
		return nil
	})
	return group.Wait()
}
