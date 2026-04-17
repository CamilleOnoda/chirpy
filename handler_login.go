package main

import (
	"encoding/json"
	"net/http"
	"time"

	auth "github.com/CamilleOnoda/chirpy/internal/auth"
	"github.com/CamilleOnoda/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	correctPassword, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !correctPassword {
		http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.jwt_secret, time.Hour)
	if err != nil {
		http.Error(w, "Could not create token.", http.StatusInternalServerError)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	const day = 24 * time.Hour
	expirationDuration := 60 * day
	expirationTime := time.Now().UTC().Add(expirationDuration)
	dbToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: expirationTime,
	})
	if err != nil {
		http.Error(w, "Error storing refresh token to the database.", http.StatusInternalServerError)
		return
	}

	responseUser := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: dbToken.Token,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseUser)
}
