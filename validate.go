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

func chirpValidater(w http.ResponseWriter, r *http.Request) {

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
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	filteredWords := []string{}
	for _, word := range strings.Split(chirp.Body, " ") {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			word = "****"
		}
		filteredWords = append(filteredWords, word)
	}
	chirp.Body = strings.Join(filteredWords, " ")

	if len(chirp.Body) > 140 {
		w.WriteHeader(400)
		result.Error = "Chirp is too long"
		result.Valid = false
	} else {
		w.WriteHeader(200)
		result.Error = "None"
		result.Valid = true
	}
	result.Clean = chirp.Body

	data, _ := json.Marshal(result)
	w.Write(data)
}
