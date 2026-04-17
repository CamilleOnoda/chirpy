package main

import (
	"encoding/json"
	"net/http"
	"time"

	auth "github.com/CamilleOnoda/chirpy/internal/auth"
	"github.com/CamilleOnoda/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashed_password, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusBadRequest)
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	responseUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseUser)
}
