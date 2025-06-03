package application

import (
	"context"
	"github.com/azaliaz/quote-service/internal/storage"
	"log/slog"
	"time"
)

//go:generate mockgen -source=service.go -destination=./mocks/service_mock.go -package=mocks

type QuoteService interface {
	AddQuote(ctx context.Context, req *AddQuoteRequest) (*AddQuoteResponse, error)
	GetQuotes(ctx context.Context, req *GetQuotesRequest) (*GetQuotesResponse, error)
	GetRandomQuote(ctx context.Context, req *GetRandomQuoteRequest) (*GetRandomQuoteResponse, error)
	GetQuotesByAuthor(ctx context.Context, req *GetQuotesByAuthorRequest) (*GetQuotesByAuthorResponse, error)
	DeleteQuote(ctx context.Context, req *DeleteQuoteRequest) (*DeleteQuoteResponse, error)
}
type Quote struct {
	ID        int64
	Author    string
	Quote     string
	CreatedAt time.Time
}
type AddQuoteRequest struct {
	Author string `json:"author"`
	Quote  string `json:"quote"`
}
type AddQuoteResponse struct {
	ID int64
}

type GetQuotesRequest struct{}
type GetQuotesResponse struct {
	Quotes []Quote
}

type GetRandomQuoteRequest struct{}
type GetRandomQuoteResponse struct {
	Quote Quote
}

type GetQuotesByAuthorRequest struct {
	Author string
}
type GetQuotesByAuthorResponse struct {
	Quotes []Quote
}

type DeleteQuoteRequest struct {
	ID int64
}
type DeleteQuoteResponse struct {
	Success bool
}

type Service struct {
	Log    *slog.Logger
	Config *Config
	DB     storage.QuoteStorage
}

func NewService(
	logger *slog.Logger,
	config *Config,
	db storage.QuoteStorage,
) *Service {
	return &Service{
		Log:    logger,
		Config: config,
		DB:     db,
	}
}

func (s *Service) Init() error {
	return nil
}

func (s *Service) Run(ctx context.Context) {

}

func (s *Service) Stop() {

}
