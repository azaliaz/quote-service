package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

//go:generate mockgen -source=service.go -destination=./mocks/service_mock.go -package=mocks

type QuoteStorage interface {
	AddQuote(ctx context.Context, quote *Quote) (int64, error)
	GetAllQuotes(ctx context.Context) ([]*Quote, error)
	GetRandomQuote(ctx context.Context) (*Quote, error)
	GetQuotesByAuthor(ctx context.Context, author string) ([]*Quote, error)
	DeleteQuote(ctx context.Context, id int64) error
}
type Quote struct {
	ID        int64
	Author    string
	Quote     string
	CreatedAt time.Time
}
type AddQuoteRequest struct {
	Author string
	Quote  string
}

func NewService(db *DB, logger *slog.Logger) *Service {
	return &Service{DB: db, logger: logger}
}

type Service struct {
	*DB
	logger *slog.Logger
}

func NewDB(config *Config, logEntry *slog.Logger) *DB {
	return &DB{
		config: config,
		log:    logEntry,
	}
}

type DB struct {
	config *Config
	log    *slog.Logger
	pool   *pgxpool.Pool
	cancel func()
}

func (r *DB) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	poolCfg, err := pgxpool.ParseConfig(r.config.dsnPostgres(r.log))
	if err != nil {
		return fmt.Errorf("error on parsing rw storage config: %w", err)
	}

	poolCfg.MaxConns = r.config.MaxOpenConns
	poolCfg.MaxConnIdleTime = r.config.ConnIdleLifetime
	poolCfg.MaxConnLifetime = r.config.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("error on creating rw storage connection pool: %w", err)
	}

	r.pool = pool

	r.log.Info("connected to postgres")
	return nil
}

func (r *DB) Run(_ context.Context) {
}

func (r *DB) Stop() {
	r.log.Info("stopping storage service")
	if r.cancel != nil {
		r.cancel()
	}
	r.pool.Close()
	r.log.Info("storage service has been stopped")
}

func (r *DB) Pool() *pgxpool.Pool {
	return r.pool
}
