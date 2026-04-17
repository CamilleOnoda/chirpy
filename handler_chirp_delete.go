package main

import (
	"net/http"

	"github.com/CamilleOnoda/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Error fetching token.", http.StatusUnauthorized)
		return
	}
	validToken, err := auth.ValidateJWT(token, cfg.jwt_secret)
	if err != nil {
		http.Error(w, "Error validating token.", http.StatusUnauthorized)
		return
	}

	chirpID := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid UUID format"))
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Couldn't retrieve chirp."))
		return
	}
	if dbChirp.UserID != validToken {
		http.Error(w, "Chirp ID and user ID do not match.", http.StatusForbidden)
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), dbChirp.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)

}
