package main

import (
	"net/http"

	"github.com/CamilleOnoda/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Error revoking token.", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
