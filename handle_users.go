package main

import (
	"encoding/json"
	"go_server/internal/auth"
	"go_server/internal/database"
	"net/http"
	"time"
)

type Response struct {
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	ID        string `json:"id"`
}
type LoginResponse struct {
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Token     string `json:"token"`
	Email     string `json:"email"`
	ID        string `json:"id"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {

	var params struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request payload", err)
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	CreateUserParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	user, err := cfg.data.CreateUser(r.Context(), CreateUserParams)
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

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	expiresIn := time.Hour
	if params.ExpiresInSeconds != 0 {
		expiresIn = time.Duration(params.ExpiresInSeconds) * time.Second
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request payload", err)
		return
	}
	if err != nil {
		respondWithError(w, 401, "Invalid authorization header", err)
		return
	}
	response, err := cfg.data.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "Failed to retrieve user", err)
		return
	}

	isValid, err := auth.CheckPasswordHash(params.Password, response.HashedPassword)
	if err != nil {
		respondWithError(w, 500, "Failed to check password", err)
		return
	}
	if !isValid {
		respondWithError(w, 401, "Invalid credentials", nil)
		return
	}
	authToken, err := auth.MakeJWT(response.ID, cfg.secret_key, expiresIn)
	user := LoginResponse{
		CreatedAt: response.CreatedAt.String(),
		UpdatedAt: response.UpdatedAt.String(),
		Token:     authToken,
		Email:     response.Email,
		ID:        response.ID.String(),
	}
	respondWithJSON(w, 200, user)
}
