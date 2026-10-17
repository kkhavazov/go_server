package main

import (
	"encoding/json"
	"fmt"
	"go_server/internal/database"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(params.Body, badWords)
	userID, err := uuid.Parse(params.UserID)

	chirp, err := cfg.data.CreateChirp(r.Context(), database.CreateChirpParams{
		Content: cleaned,
		UserID:  userID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating chirp: %v", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		ID        string `json:"id"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        chirp.ID.String(),
		Body:      chirp.Content,
		UserID:    chirp.UserID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (cfg *apiConfig) handlerChirpsRead(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.data.GetChirps(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving chirps: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]struct {
		ID        string `json:"id"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}, len(chirps))

	for i, chirp := range chirps {
		response[i] = struct {
			ID        string `json:"id"`
			Body      string `json:"body"`
			UserID    string `json:"user_id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}{
			ID:        chirp.ID.String(),
			Body:      chirp.Content,
			UserID:    chirp.UserID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
		}
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerChirpsReadByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	chirpID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid chirp ID: %v", err), http.StatusBadRequest)
		return
	}

	chirp, err := cfg.data.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving chirp: %v", err), http.StatusNotFound)
		return
	}

	response := struct {
		ID        string `json:"id"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        chirp.ID.String(),
		Body:      chirp.Content,
		UserID:    chirp.UserID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
	}

	respondWithJSON(w, http.StatusOK, response)
}
