package main

import (
	"encoding/json"
	"net/http"

	"go_server/internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleWebhooksCreate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	type data struct {
		UserID string `json:"user_id"`
	}
	var params struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	api_key, err := auth.GetAPIKey(r.Header)

	if err != nil {
		respondWithError(w, 401, "Wrong API Key", err)
		return
	}
	if api_key != cfg.polka_key {
		respondWithError(w, 401, "Wrong API Key", err)
		return
	}

	id, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not create uuid", err)
		return
	}
	if params.Event == "user.upgraded" {

		err = cfg.data.ChangeToRed(r.Context(), id)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not upgared user", err)
			return
		}
	}
	respondWithJSON(w, 204, nil)

}
