package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortDir := r.URL.Query().Get("sort")

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	if authorId != "" {
		id, _ := strconv.Atoi(authorId)
		for _, dbChirp := range dbChirps {
			if dbChirp.AuthorId == id {
				chirps = append(chirps, Chirp{
					ID:       dbChirp.ID,
					Body:     dbChirp.Body,
					AuthorId: dbChirp.AuthorId,
				})
			}
		}
	} else {
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, Chirp{
				ID:       dbChirp.ID,
				Body:     dbChirp.Body,
				AuthorId: dbChirp.AuthorId,
			})
		}
	}

	if sortDir == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpRetrieve(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("chirpID"))

	chirp, err := cfg.DB.GetChirp(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp with that ID")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
