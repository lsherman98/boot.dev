package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Password string `json:"password"`
        Email string `json:"email"`
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
        return 
    }

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
        respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid or expired token: %v", err),)
        return
    }

    userId, _ := token.Claims.GetSubject()
    user, err := cfg.DB.UpdateUser(params.Email, params.Password, userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}

    respondWithJSON(w, http.StatusOK, User{
		ID:   user.ID,
		Email: user.Email,
        IsChirpyRed: user.IsChirpyRed,
	})
}