package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)



func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
    authorizationHeader := r.Header.Get("Authorization")
    if authorizationHeader == "" {
        respondWithError(w, http.StatusUnauthorized, "Authorization header missing")
        return
    }
    refreshTokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")

    userId, err := cfg.DB.ValidateRefreshToken(refreshTokenString)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Refresh token doese not exist in DB or is expired - %v", err))
        return
    }

    secondsInHour := 3600
	expirationTime := time.Now().Add(time.Second * time.Duration(secondsInHour))
	claims := jwt.RegisteredClaims{
        Issuer: "chirpy", 
        IssuedAt: jwt.NewNumericDate(time.Now()), 
        ExpiresAt: jwt.NewNumericDate(expirationTime),
        Subject: strconv.Itoa(userId),
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

    respondWithJSON(w, http.StatusOK, struct{Token string `json:"token"`}{
        Token: signedToken,
    })
}