package main

import (
	"encoding/json"
	"net/http"

	"github.com/CamilleOnoda/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgrade(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if apiKey != cfg.polka_key {
		http.Error(w, "Wrong API Key", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error decodng JSON", http.StatusBadRequest)
		return
	}

	if user.Event == "user.upgraded" {
		_, err := cfg.db.UpgradeByID(r.Context(), user.Data.UserID)
		if err != nil {
			http.Error(w, "Error upgrading user", http.StatusNotFound)
		}
	}
	w.WriteHeader(http.StatusNoContent)

}
