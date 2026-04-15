package main

import (
	"encoding/json"
	"net/http"

	auth "github.com/CamilleOnoda/chirpy/internal/auth"
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
	}
	correctPassword, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !correctPassword {
		http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
	}
	responseUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseUser)
}
