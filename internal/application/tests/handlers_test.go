package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/azaliaz/quote-service/internal/application"
	"github.com/azaliaz/quote-service/internal/storage"
	"github.com/azaliaz/quote-service/internal/storage/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAddQuote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		req  *application.AddQuoteRequest
		mock func(m *mocks.MockQuoteStorage)
		want *application.AddQuoteResponse
		err  string
	}{
		{
			name: "success",
			req:  &application.AddQuoteRequest{Author: "Author", Quote: "Quote"},
			mock: func(m *mocks.MockQuoteStorage) {
				m.EXPECT().
					AddQuote(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			want: &application.AddQuoteResponse{ID: 1},
			err:  "",
		},
		{
			name: "empty author",
			req:  &application.AddQuoteRequest{Author: "", Quote: "Quote"},
			mock: nil,
			want: nil,
			err:  "author and quote cannot be empty",
		},
		{
			name: "storage error",
			req:  &application.AddQuoteRequest{Author: "Author", Quote: "Quote"},
			mock: func(m *mocks.MockQuoteStorage) {
				m.EXPECT().
					AddQuote(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("db error"))
			},
			want: nil,
			err:  "failed to add quote: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockQuoteStorage(ctrl)
			if tt.mock != nil {
				tt.mock(mockStorage)
			}

			svc := &application.Service{
				DB:  mockStorage,
				Log: newTestLogger(),
			}

			resp, err := svc.AddQuote(context.Background(), tt.req)
			if tt.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestGetQuotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockQuoteStorage(ctrl)
	mockStorage.EXPECT().
		GetAllQuotes(gomock.Any()).
		Return([]*storage.Quote{
			{ID: 1, Author: "A", Quote: "Q1", CreatedAt: time.Now()},
			{ID: 2, Author: "B", Quote: "Q2", CreatedAt: time.Now()},
		}, nil)

	svc := &application.Service{
		DB:  mockStorage,
		Log: newTestLogger(),
	}

	resp, err := svc.GetQuotes(context.Background(), &application.GetQuotesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Quotes, 2)
	assert.Equal(t, "A", resp.Quotes[0].Author)
	assert.Equal(t, "Q1", resp.Quotes[0].Quote)
}

func TestGetRandomQuote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockQuoteStorage(ctrl)
	mockStorage.EXPECT().
		GetRandomQuote(gomock.Any()).
		Return(&storage.Quote{
			ID:        1,
			Author:    "Author",
			Quote:     "Random quote",
			CreatedAt: time.Now(),
		}, nil)

	svc := &application.Service{
		DB:  mockStorage,
		Log: newTestLogger(),
	}

	resp, err := svc.GetRandomQuote(context.Background(), &application.GetRandomQuoteRequest{})
	require.NoError(t, err)
	assert.Equal(t, "Author", resp.Quote.Author)
	assert.Equal(t, "Random quote", resp.Quote.Quote)
}

func TestGetQuotesByAuthor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()

	tests := []struct {
		name   string
		req    *application.GetQuotesByAuthorRequest
		mock   func(m *mocks.MockQuoteStorage)
		want   *application.GetQuotesByAuthorResponse
		errMsg string
	}{
		{
			name: "success",
			req:  &application.GetQuotesByAuthorRequest{Author: "Author"},
			mock: func(m *mocks.MockQuoteStorage) {
				m.EXPECT().
					GetQuotesByAuthor(gomock.Any(), "Author").
					Return([]*storage.Quote{
						{
							ID:        1,
							Author:    "Author",
							Quote:     "Quote 1",
							CreatedAt: now,
						},
					}, nil)
			},
			want: &application.GetQuotesByAuthorResponse{
				Quotes: []application.Quote{
					{
						ID:     1,
						Author: "Author",
						Quote:  "Quote 1",
						// CreatedAt не сравниваем напрямую
					},
				},
			},
			errMsg: "",
		},
		{
			name:   "empty author",
			req:    &application.GetQuotesByAuthorRequest{Author: ""},
			mock:   nil,
			want:   nil,
			errMsg: "author parameter is required",
		},
		{
			name: "storage error",
			req:  &application.GetQuotesByAuthorRequest{Author: "Author"},
			mock: func(m *mocks.MockQuoteStorage) {
				m.EXPECT().
					GetQuotesByAuthor(gomock.Any(), "Author").
					Return(nil, errors.New("db error"))
			},
			want:   nil,
			errMsg: "failed to get quotes by author: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockQuoteStorage(ctrl)
			if tt.mock != nil {
				tt.mock(mockStorage)
			}

			svc := &application.Service{
				DB:  mockStorage,
				Log: newTestLogger(),
			}

			resp, err := svc.GetQuotesByAuthor(context.Background(), tt.req)
			if tt.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Len(t, resp.Quotes, len(tt.want.Quotes))

				for i, gotQuote := range resp.Quotes {
					wantQuote := tt.want.Quotes[i]
					assert.Equal(t, wantQuote.ID, gotQuote.ID)
					assert.Equal(t, wantQuote.Author, gotQuote.Author)
					assert.Equal(t, wantQuote.Quote, gotQuote.Quote)
					// Проверяем CreatedAt с допуском в 1 секунду
					assert.WithinDuration(t, now, gotQuote.CreatedAt, time.Second)
				}
			}
		})
	}
}

func TestDeleteQuote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name   string
		req    *application.DeleteQuoteRequest
		mock   func(m *mocks.MockQuoteStorage)
		want   *application.DeleteQuoteResponse
		errMsg string
	}{
		{
			name: "success",
			req:  &application.DeleteQuoteRequest{ID: 1},
			mock: func(m *mocks.MockQuoteStorage) {
				m.EXPECT().
					DeleteQuote(gomock.Any(), int64(1)).
					Return(nil)
			},
			want:   &application.DeleteQuoteResponse{Success: true},
			errMsg: "",
		},
		{
			name:   "empty id",
			req:    &application.DeleteQuoteRequest{ID: 0},
			mock:   nil,
			want:   nil,
			errMsg: "id parameter is required",
		},
		{
			name: "storage error",
			req:  &application.DeleteQuoteRequest{ID: 1},
			mock: func(m *mocks.MockQuoteStorage) {
				m.EXPECT().
					DeleteQuote(gomock.Any(), int64(1)).
					Return(errors.New("db error"))
			},
			want:   &application.DeleteQuoteResponse{Success: false},
			errMsg: "failed to delete quote: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockQuoteStorage(ctrl)
			if tt.mock != nil {
				tt.mock(mockStorage)
			}

			svc := &application.Service{
				DB:  mockStorage,
				Log: newTestLogger(),
			}

			resp, err := svc.DeleteQuote(context.Background(), tt.req)
			if tt.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				if tt.want == nil {
					assert.Nil(t, resp) // ожидаем nil, если want == nil
				} else {
					require.NotNil(t, resp)
					assert.False(t, resp.Success)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}

		})
	}
}
