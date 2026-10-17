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
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Email        string `json:"email"`
	ID           string `json:"id"`
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	expiresInJWT := time.Hour
	exporesInAT := time.Hour * 24 * 60
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Invalid request payload", err)
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
	newToken := auth.MakeRefreshToken()
	CreateTokenParams := database.AddTokenParams{
		UserID:    response.ID,
		Token:     newToken,
		ExpiresAt: time.Now().Add(exporesInAT),
	}
	_, err = cfg.data.AddToken(r.Context(), CreateTokenParams)
	if err != nil {
		respondWithError(w, 500, "Failed to create token", err)
		return
	}
	authToken, err := auth.MakeJWT(response.ID, cfg.secret_key, expiresInJWT)
	if err != nil {
		respondWithError(w, 500, "Failed to create token", err)
		return
	}
	user := LoginResponse{
		CreatedAt:    response.CreatedAt.String(),
		UpdatedAt:    response.UpdatedAt.String(),
		Token:        authToken,
		RefreshToken: newToken,
		Email:        response.Email,
		ID:           response.ID.String(),
	}
	respondWithJSON(w, 200, user)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	ValidatedUserID, err := auth.ValidateJWT(token, cfg.secret_key)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	hashed_pass, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not hush password", err)
		return
	}

	response, err := cfg.data.UpdatePassword(r.Context(), database.UpdatePasswordParams{HashedPassword: hashed_pass, Email: params.Email, ID: ValidatedUserID})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email", err)
		return
	}
	result := Response{
		CreatedAt: response.CreatedAt.String(),
		UpdatedAt: response.UpdatedAt.String(),
		Email:     response.Email,
		ID:        response.ID.String(),
	}
	respondWithJSON(w, 200, result)
}
