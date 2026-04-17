package main

import (
	"encoding/json"
	"net/http"

	"github.com/CamilleOnoda/chirpy/internal/auth"
	"github.com/CamilleOnoda/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUsersPut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error decoding JSON"))
		return
	}
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
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hasing password.", http.StatusInternalServerError)
	}
	updateUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          user.Email,
		HashedPassword: hashedPassword,
		ID:             validToken,
	})
	if err != nil {
		http.Error(w, "Error updating user", http.StatusUnauthorized)
	}
	response := User{
		Email: updateUser.Email,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}
