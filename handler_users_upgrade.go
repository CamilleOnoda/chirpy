package main

import (
	"encoding/json"
	"net/http"

	"github.com/CamilleOnoda/chirpy/internal/database"
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

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error decodng JSON", http.StatusBadRequest)
		return
	}

	if user.Event == "user.upgraded" {
		_, err := cfg.db.UpgradeByID(r.Context(), database.UpgradeByIDParams{
			IsChirpyRed: true,
			ID:          user.Data.UserID,
		})
		if err != nil {
			http.Error(w, "Error upgrading user", http.StatusNotFound)
		}
	}
	w.WriteHeader(http.StatusNoContent)

}
