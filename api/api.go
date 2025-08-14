package api

import (
	"encoding/json"
	"net/http"
)

type Card struct {
	ListingTitle string
	Price        float64
	ImageURL     string
}

type CardAPIParams struct {
	SearchCrit string
}

type CardSearchResponse struct {
	StatusCode   int
	AveragePrice float64
	Cards        []Card
}

type ErrorResponse struct {
	StatusCode int
	Message    string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := ErrorResponse{
		StatusCode: code,
		Message:    message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

var RequestErrorHandler = func(w http.ResponseWriter, err error) {
	writeError(w, err.Error(), http.StatusBadRequest)
}

var InternalErrorHandler = func(w http.ResponseWriter) {
	writeError(w, "Internal Error", http.StatusInternalServerError)
}
