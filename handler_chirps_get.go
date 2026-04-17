package main

import (
	"encoding/json"
	"net/http"

	"github.com/CamilleOnoda/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var dbChirps []database.Chirp
	var err error

	authorID := r.URL.Query().Get("author_id")
	if authorID == "" {
		dbChirps, err = cfg.db.GetAllChirps(r.Context())
	} else {
		id, parseErr := uuid.Parse(authorID)
		if parseErr != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}
		dbChirps, err = cfg.db.GetChirpsByAuthor(r.Context(), id)
	}
	if err != nil {
		http.Error(w, "Couldn't retrieve chirps.", http.StatusInternalServerError)
	}
	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}
	data, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
