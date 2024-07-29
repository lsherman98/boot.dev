package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)


func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
    chirpId, _ := strconv.Atoi(r.PathValue("chirpID"))


	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header missing")
		return
	}
	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid or expired token: %v", err))
		return
	}
	userId, _ := token.Claims.GetSubject()

	err = cfg.DB.DeleteChirp(chirpId, userId)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Couldn't delete chirp")
		return
	}
    w.WriteHeader(http.StatusNoContent)

}