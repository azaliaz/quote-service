package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/azaliaz/quote-service/internal/application"
	"github.com/azaliaz/quote-service/internal/application/mocks"
	"github.com/azaliaz/quote-service/internal/facade/rest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAPI(t *testing.T) (*rest.Service, *mocks.MockQuoteService) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockQuoteService(ctrl)
	api := &rest.Service{App: mockSvc}
	t.Cleanup(ctrl.Finish)
	return api, mockSvc
}

func TestHandleQuotes(t *testing.T) {
	api, mockSvc := newTestAPI(t)

	t.Run("GET /quotes", func(t *testing.T) {
		mockSvc.EXPECT().
			GetQuotes(gomock.Any(), gomock.Any()).
			Return(&application.GetQuotesResponse{Quotes: []application.Quote{}}, nil)

		req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
		rr := httptest.NewRecorder()

		api.HandleQuotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("POST /quotes", func(t *testing.T) {
		reqBody := `{"author":"Author","quote":"Quote"}`
		mockSvc.EXPECT().
			AddQuote(gomock.Any(), &application.AddQuoteRequest{Author: "Author", Quote: "Quote"}).
			Return(&application.AddQuoteResponse{ID: 42}, nil)

		req := httptest.NewRequest(http.MethodPost, "/quotes", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		api.HandleQuotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp map[string]int64
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, int64(42), resp["id"])
	})

	t.Run("Unsupported method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/quotes", nil)
		rr := httptest.NewRecorder()

		api.HandleQuotes(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestHandleRandomQuote(t *testing.T) {
	api, mockSvc := newTestAPI(t)

	t.Run("GET /quotes/random success", func(t *testing.T) {
		mockSvc.EXPECT().
			GetRandomQuote(gomock.Any(), gomock.Any()).
			Return(&application.GetRandomQuoteResponse{
				Quote: application.Quote{ID: 1, Author: "A", Quote: "Q"},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/quotes/random", nil)
		rr := httptest.NewRecorder()

		api.HandleRandomQuote(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"Author": "A"`)
	})

	t.Run("Non-GET method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/quotes/random", nil)
		rr := httptest.NewRecorder()

		api.HandleRandomQuote(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("App error", func(t *testing.T) {
		mockSvc.EXPECT().
			GetRandomQuote(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("fail"))

		req := httptest.NewRequest(http.MethodGet, "/quotes/random", nil)
		rr := httptest.NewRecorder()

		api.HandleRandomQuote(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to get random quote")
	})
}

func TestHandleQuoteByID(t *testing.T) {
	api, mockSvc := newTestAPI(t)

	t.Run("DELETE /quotes/{id} success", func(t *testing.T) {
		mockSvc.EXPECT().
			DeleteQuote(gomock.Any(), &application.DeleteQuoteRequest{ID: 123}).
			Return(&application.DeleteQuoteResponse{Success: true}, nil)

		req := httptest.NewRequest(http.MethodDelete, "/quotes/123", nil)
		rr := httptest.NewRecorder()

		api.HandleQuoteByID(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/quotes/123", nil)
		rr := httptest.NewRecorder()

		api.HandleQuoteByID(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/quotes/123/extra", nil)
		rr := httptest.NewRecorder()

		api.HandleQuoteByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/quotes/abc", nil)
		rr := httptest.NewRecorder()

		api.HandleQuoteByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid quote ID")
	})

	t.Run("Delete error", func(t *testing.T) {
		mockSvc.EXPECT().
			DeleteQuote(gomock.Any(), &application.DeleteQuoteRequest{ID: 123}).
			Return(nil, errors.New("fail"))

		req := httptest.NewRequest(http.MethodDelete, "/quotes/123", nil)
		rr := httptest.NewRecorder()

		api.HandleQuoteByID(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to delete quote")
	})

	t.Run("Delete not success", func(t *testing.T) {
		mockSvc.EXPECT().
			DeleteQuote(gomock.Any(), &application.DeleteQuoteRequest{ID: 123}).
			Return(&application.DeleteQuoteResponse{Success: false}, nil)

		req := httptest.NewRequest(http.MethodDelete, "/quotes/123", nil)
		rr := httptest.NewRecorder()

		api.HandleQuoteByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Quote not found")
	})
}

func TestAddQuote(t *testing.T) {
	api, mockSvc := newTestAPI(t)

	t.Run("Valid add quote", func(t *testing.T) {
		reqBody := `{"author":"Author","quote":"Quote"}`
		mockSvc.EXPECT().
			AddQuote(gomock.Any(), &application.AddQuoteRequest{Author: "Author", Quote: "Quote"}).
			Return(&application.AddQuoteResponse{ID: 42}, nil)

		req := httptest.NewRequest(http.MethodPost, "/quotes", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		api.AddQuote(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp map[string]int64
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, int64(42), resp["id"])
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/quotes", strings.NewReader("{invalid json"))
		rr := httptest.NewRecorder()

		api.AddQuote(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid JSON body")
	})

	t.Run("Missing fields", func(t *testing.T) {
		reqBody := `{"author":"","quote":""}`
		req := httptest.NewRequest(http.MethodPost, "/quotes", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		api.AddQuote(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "author and quote fields are required")
	})

	t.Run("App error", func(t *testing.T) {
		reqBody := `{"author":"Author","quote":"Quote"}`
		mockSvc.EXPECT().
			AddQuote(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("fail"))

		req := httptest.NewRequest(http.MethodPost, "/quotes", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		api.AddQuote(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to add quote")
	})
}

func TestGetQuotes(t *testing.T) {
	api, mockSvc := newTestAPI(t)

	t.Run("Get all quotes", func(t *testing.T) {
		mockSvc.EXPECT().
			GetQuotes(gomock.Any(), gomock.Any()).
			Return(&application.GetQuotesResponse{
				Quotes: []application.Quote{
					{ID: 1, Author: "A", Quote: "Q1"},
					{ID: 2, Author: "B", Quote: "Q2"},
				},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
		rr := httptest.NewRecorder()

		api.GetQuotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"Author": "A"`)
	})

	t.Run("Get quotes by author", func(t *testing.T) {
		mockSvc.EXPECT().
			GetQuotesByAuthor(gomock.Any(), &application.GetQuotesByAuthorRequest{Author: "Author"}).
			Return(&application.GetQuotesByAuthorResponse{
				Quotes: []application.Quote{
					{ID: 1, Author: "Author", Quote: "Quote 1"},
				},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/quotes?author=Author", nil)
		rr := httptest.NewRecorder()

		api.GetQuotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"Author": "Author"`)
	})

	t.Run("App error", func(t *testing.T) {
		mockSvc.EXPECT().
			GetQuotes(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("fail"))

		req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
		rr := httptest.NewRecorder()

		api.GetQuotes(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to get quotes")
	})
}
