package main

import (
	"go_server/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 404, "Invalid token format", err)
		return
	}
	err = cfg.data.RevokeToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 404, "Invalid token", err)
		return
	}

	respondWithJSON(w, 204, nil)

}
