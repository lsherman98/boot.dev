package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type loginResponse struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
        RefreshToken string `json:"refresh_token"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.DB.AuthenticateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Login failed")
		return
	}

    refreshToken, err := cfg.DB.GenerateRefreshToken(user.ID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
        return
    }

	secondsInHour := 3600
	expirationTime := time.Now().Add(time.Second * time.Duration(secondsInHour))
	claims := jwt.RegisteredClaims{
        Issuer: "chirpy", 
        IssuedAt: jwt.NewNumericDate(time.Now()), 
        ExpiresAt: jwt.NewNumericDate(expirationTime),
        Subject: strconv.Itoa(user.ID),
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:    user.ID,
		Email: user.Email,
		Token: signedToken,
        RefreshToken: refreshToken,
		IsChirpyRed: user.IsChirpyRed,
	})
}
