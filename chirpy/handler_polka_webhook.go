package main

import (
	"encoding/json"
	"net/http"
	"strings"
)


func (cfg *apiConfig) handlerPolkaWebHook(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Event string `json:"event"`
        Data struct{
            UserId int `json:"user_id"`
        } `json:"data"`
    }

    apiKeyHeader := r.Header.Get("Authorization")
    apiKey := strings.TrimPrefix(apiKeyHeader, "ApiKey ")
    if apiKeyHeader == "" || apiKey != cfg.ApiKey {
        respondWithError(w, http.StatusUnauthorized, "Authorization Failed")
        return
    }

    decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

    if params.Event != "user.upgraded" {
        w.WriteHeader(http.StatusNoContent)
        return 
    }

    err = cfg.DB.UpgradeUser(params.Data.UserId)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
    }

    w.WriteHeader(http.StatusNoContent)
}
    