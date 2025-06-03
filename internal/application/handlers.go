package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/azaliaz/quote-service/internal/storage"

	"time"
)

func toAppQuote(sq *storage.Quote) Quote {
	return Quote{
		ID:        sq.ID,
		Author:    sq.Author,
		Quote:     sq.Quote,
		CreatedAt: sq.CreatedAt,
	}
}

func (s *Service) AddQuote(ctx context.Context, req *AddQuoteRequest) (*AddQuoteResponse, error) {
	if req.Author == "" || req.Quote == "" {
		return nil, errors.New("author and quote cannot be empty")
	}

	quote := storage.Quote{
		Author:    req.Author,
		Quote:     req.Quote,
		CreatedAt: time.Now(),
	}

	id, err := s.DB.AddQuote(ctx, &quote)
	if err != nil {
		s.Log.Error("failed to add quote", "error", err)
		return nil, fmt.Errorf("failed to add quote: %w", err)
	}

	return &AddQuoteResponse{ID: id}, nil
}

func (s *Service) GetQuotes(ctx context.Context, req *GetQuotesRequest) (*GetQuotesResponse, error) {
	quotes, err := s.DB.GetAllQuotes(ctx)
	if err != nil {
		s.Log.Error("failed to get quotes", "error", err)
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	result := make([]Quote, 0, len(quotes))

	for _, q := range quotes {
		result = append(result, Quote{
			ID:        q.ID,
			Author:    q.Author,
			Quote:     q.Quote,
			CreatedAt: q.CreatedAt,
		})
	}

	return &GetQuotesResponse{Quotes: result}, nil
}

func (s *Service) GetRandomQuote(ctx context.Context, req *GetRandomQuoteRequest) (*GetRandomQuoteResponse, error) {
	storageQuote, err := s.DB.GetRandomQuote(ctx)
	if err != nil {
		s.Log.Error("failed to get random quote", "error", err)
		return nil, fmt.Errorf("failed to get random quote: %w", err)
	}

	return &GetRandomQuoteResponse{Quote: toAppQuote(storageQuote)}, nil
}

func (s *Service) GetQuotesByAuthor(ctx context.Context, req *GetQuotesByAuthorRequest) (*GetQuotesByAuthorResponse, error) {
	if req.Author == "" {
		return nil, errors.New("author parameter is required")
	}

	quotes, err := s.DB.GetQuotesByAuthor(ctx, req.Author)
	if err != nil {
		s.Log.Error("failed to get quotes by author", "author", req.Author, "error", err)
		return nil, fmt.Errorf("failed to get quotes by author: %w", err)
	}

	result := make([]Quote, 0, len(quotes))
	for _, q := range quotes {
		result = append(result, toAppQuote(q))
	}

	return &GetQuotesByAuthorResponse{Quotes: result}, nil
}

func (s *Service) DeleteQuote(ctx context.Context, req *DeleteQuoteRequest) (*DeleteQuoteResponse, error) {
	if req.ID == 0 {
		return nil, errors.New("id parameter is required")
	}

	err := s.DB.DeleteQuote(ctx, req.ID)
	if err != nil {
		s.Log.Error("failed to delete quote", "id", req.ID, "error", err)
		return &DeleteQuoteResponse{Success: false}, fmt.Errorf("failed to delete quote: %w", err)
	}

	return &DeleteQuoteResponse{Success: true}, nil
}
