package main

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
	ID        string `json:"id"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {

	var params struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request payload", err)
		return
	}
	user, err := cfg.data.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "Failed to create user", err)
		return
	}
	response := Response{
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
		ID:        user.ID.String(),
	}
	respondWithJSON(w, 201, response)
}
