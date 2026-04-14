package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	auth "github.com/CamilleOnoda/chirpy/internal/auth"
	"github.com/CamilleOnoda/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Couldn't retrieve chirps."))
		return
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

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirpID := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid UUID format"))
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Couldn't retrieve chirp."))
		return
	}
	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func validateChirp(chirpBody string) (string, error) {
	if len(chirpBody) > 140 {
		err := fmt.Errorf("Chirp is too long")
		return "", err
	}
	cleaned := cleanChirpBody(chirpBody)
	return cleaned, nil
}

func cleanChirpBody(body string) string {
	notAllowedWords := map[string]bool{"kerfuffle": true, "sharbert": true, "fornax": true}
	words := strings.Split(body, " ")
	for i, word := range words {
		if notAllowedWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	cfg := &apiConfig{
		db:       database.New(db),
		platform: os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("./internal/static/")))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	},
	)

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1>"+
			"<p>Chirpy has been visited %d times!</p></body></html>",
			cfg.fileserverHits.Load())))
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if cfg.platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Reset is only allowed in dev environment."))
			return
		}

		cfg.fileserverHits.Store(0)
		err := cfg.db.DeleteAllUsers(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to reset the database: " + err.Error()))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Counter reset to %d", cfg.fileserverHits.Load())))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var chirp struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error decoding JSON"))
			return
		}
		cleanedChirp, err := validateChirp(chirp.Body)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   cleanedChirp,
			UserID: chirp.UserID,
		})
		if err != nil {
			http.Error(w, "Error creating chirp", http.StatusBadRequest)
			return
		}

		responseChirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responseChirp)

	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirpByID)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Server started on port: %v", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
