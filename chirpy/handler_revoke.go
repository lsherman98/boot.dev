package main

import (
	"fmt"
	"net/http"
	"strings"
)



func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
    authorizationHeader := r.Header.Get("Authorization")
    if authorizationHeader == "" {
        respondWithError(w, http.StatusUnauthorized, "Authorization header missing")
        return
    }
    refreshTokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")

    err := cfg.DB.RevokeRefreshToken(refreshTokenString)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldnt revoke token: %v", err))
        return
    }

	w.WriteHeader(http.StatusNoContent)
}