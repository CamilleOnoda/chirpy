package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/CamilleOnoda/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	dbUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Error getting refresh token from the database.", http.StatusUnauthorized)
		return
	}
	newAccess, err := auth.MakeJWT(dbUser.ID, cfg.jwt_secret, time.Hour)
	if err != nil {
		http.Error(w, "Error creating a new access token.", http.StatusUnauthorized)
		return
	}

	tokenResposnse := response{
		Token: newAccess,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResposnse)
}
