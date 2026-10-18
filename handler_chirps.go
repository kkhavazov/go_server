package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"go_server/internal/auth"
	"go_server/internal/database"
	"net/http"
	"slices"

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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header", err)
		return
	}
	ValidatedUserID, err := auth.ValidateJWT(token, cfg.secret_key)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	chirp, err := cfg.data.CreateChirp(r.Context(), database.CreateChirpParams{
		Content: cleaned,
		UserID:  ValidatedUserID,
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
	var chirps []database.Chirp
	var err error
	queryParams := r.URL.Query()
	author_id := queryParams.Get("author_id")
	sort := queryParams.Get("sort")
	sort_type := "asc"
	if sort == "desc" {
		sort_type = "desc"
	}
	if author_id != "" {
		parsed_id, err := uuid.Parse(author_id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error converting ID: %v", err), http.StatusInternalServerError)
			return
		}
		chirps, err = cfg.data.GetChirpsByAuthor(r.Context(), parsed_id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving chirps: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		chirps, err = cfg.data.GetChirps(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving chirps: %v", err), http.StatusInternalServerError)
			return
		}
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
	if sort_type == "desc" {
		slices.SortFunc(response, func(a, b struct {
			ID        string `json:"id"`
			Body      string `json:"body"`
			UserID    string `json:"user_id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}) int {
			return cmp.Compare(b.CreatedAt, a.CreatedAt)
		})
	} else {
		slices.SortFunc(response, func(a, b struct {
			ID        string `json:"id"`
			Body      string `json:"body"`
			UserID    string `json:"user_id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}) int {
			return cmp.Compare(a.CreatedAt, b.CreatedAt)
		})
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

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid chirp ID: %v", err), http.StatusBadRequest)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header", err)
		return
	}
	ValidatedUserID, err := auth.ValidateJWT(token, cfg.secret_key)
	err = cfg.data.DeleteChirps(r.Context(), database.DeleteChirpsParams{
		ID:     chirpID,
		UserID: ValidatedUserID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid chirp ID: %v", err), 403)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
