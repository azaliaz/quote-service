package rest

import (
	"encoding/json"
	"github.com/azaliaz/quote-service/internal/application"
	"net/http"
	"strconv"
	"strings"
)

const (
	expectedPartsLength = 3
)

func (api *Service) HandleQuotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.GetQuotes(w, r)
	case http.MethodPost:
		api.AddQuote(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *Service) HandleRandomQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := api.App.GetRandomQuote(r.Context(), &application.GetRandomQuoteRequest{})
	if err != nil {
		http.Error(w, "Failed to get random quote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp.Quote)
}

func (api *Service) HandleQuoteByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != expectedPartsLength {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid quote ID", http.StatusBadRequest)
		return
	}

	resp, err := api.App.DeleteQuote(r.Context(), &application.DeleteQuoteRequest{ID: id})
	if err != nil {
		http.Error(w, "Failed to delete quote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !resp.Success {
		http.Error(w, "Quote not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *Service) AddQuote(w http.ResponseWriter, r *http.Request) {
	var req application.AddQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Author == "" || req.Quote == "" {
		http.Error(w, "author and quote fields are required", http.StatusBadRequest)
		return
	}

	resp, err := api.App.AddQuote(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to add quote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]int64{"id": resp.ID})
}

func (api *Service) GetQuotes(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")

	if author != "" {
		resp, err := api.App.GetQuotesByAuthor(r.Context(), &application.GetQuotesByAuthorRequest{Author: author})
		if err != nil {
			http.Error(w, "Failed to get quotes by author: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp.Quotes)
		return
	}

	resp, err := api.App.GetQuotes(r.Context(), &application.GetQuotesRequest{})
	if err != nil {
		http.Error(w, "Failed to get quotes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp.Quotes)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON: "+err.Error(), http.StatusInternalServerError)
	}
}
