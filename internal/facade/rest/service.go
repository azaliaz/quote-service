package rest

import (
	"context"
	"fmt"
	"github.com/azaliaz/quote-service/internal/application"
	"log/slog"
	"net/http"

	"time"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	idleTimeout    = 60 * time.Second
	contextTimeout = 5 * time.Second
)

type Service struct {
	Log    *slog.Logger
	Config *Config
	App    application.QuoteService // с заглавной буквы
	Server *http.Server
}

func NewAPI(logEntry *slog.Logger, config *Config, app application.QuoteService) *Service {
	return &Service{
		Log:    logEntry,
		Config: config,
		App:    app,
	}
}

func (api *Service) Init() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/quotes", api.HandleQuotes)
	mux.HandleFunc("/quotes/random", api.HandleRandomQuote)
	mux.HandleFunc("/quotes/", api.HandleQuoteByID)
	addr := fmt.Sprintf(":%d", api.Config.Port)
	api.Server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	api.Log.Info("HTTP server initialized", "addr", addr)
	return nil
}

func (api *Service) Run(ctx context.Context) {
	api.Log.Info("starting HTTP server", "addr", api.Server.Addr)
	if err := api.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.Log.Error("HTTP server error", "error", err)
	}
}

func (api *Service) Stop() {
	api.Log.Info("stopping HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	if err := api.Server.Shutdown(ctx); err != nil {
		api.Log.Error("failed to shutdown HTTP server", "error", err)
	}
}
