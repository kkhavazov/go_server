package main

import (
	"database/sql"
	"fmt"
	"go_server/internal/auth"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type TokenResponse struct {
	UserID    uuid.UUID    `json:"user_id"`
	ExpiresAt time.Time    `json:"expired_at"`
	RevokedAt sql.NullTime `json:"revoked_at"`
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type Token struct {
		Token string `json:"token"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, "Invalid token format", err)
		return
	}
	response, err := cfg.data.GetToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 400, "Invalid token", err)
		return
	}
	token_data := TokenResponse{
		UserID:    response.UserID,
		ExpiresAt: response.ExpiresAt,
		RevokedAt: response.RevokedAt,
	}
	if time.Now().After(token_data.ExpiresAt) {
		respondWithError(w, 401, "Expired key", fmt.Errorf(""))
		return
	}
	if token_data.RevokedAt.Valid {
		respondWithError(w, 401, "Expired key", fmt.Errorf(""))
		return
	}
	token_res, err := auth.MakeJWT(token_data.UserID, cfg.secret_key, time.Hour)
	if err != nil {
		respondWithError(w, 400, "Invalid token", err)
		return
	}
	result := Token{
		Token: token_res,
	}
	respondWithJSON(w, 200, result)
}
