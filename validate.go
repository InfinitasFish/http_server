package main

import (
	"net/http"
	"encoding/json"
	"strings"
	)

type chirpBody struct {
		Body string `json:"body"`
	}

type chirpResult struct {
	Error string `json:"error"`
	Valid bool `json:"valid"`
	Clean string `json:"cleaned_body"`
}

func chirpValidateHandler(w http.ResponseWriter, r *http.Request) {
	chirp := chirpBody{}
	result := chirpResult{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&chirp)
	if err != nil {
		result.Error = "Som ting went wong"
		result.Valid = false
		result.Clean = "None"
		data, _ := json.Marshal(result)
	
		w.WriteHeader(500)
		w.Write(data)
		return
    }

	// cleaning chirp text slowly O(n)
	// some "profane" words for debug, map[string] for fast checking
	chirp.Body = filterChirp(chirp.Body)
	isValid := validateChirp(chirp.Body)
	if isValid {
		w.WriteHeader(200)
		result.Error = "None"
		result.Valid = true
	} else {
		w.WriteHeader(400)
		result.Error = "Chirp is too long"
		result.Valid = false
	}

	data, _ := json.Marshal(result)
	w.Write(data)
}

func filterChirp(chirpBody string) string {
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	filteredWords := []string{}
	for _, word := range strings.Split(chirpBody, " ") {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			word = "****"
		}
		filteredWords = append(filteredWords, word)
	}
	chirpBody = strings.Join(filteredWords, " ")

	return chirpBody
}

func validateChirp(chirpBody string) bool {
	if len(chirpBody) > 140 {
		return false
	} else {
		return true
	}
}
