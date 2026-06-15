package main

import (
	"fmt"
	"net/http"
	)


func main() {
	root := http.Dir(".")
	fmt.Println(root)

	// ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request against a list 
	// of registered patterns and calls the handler for the pattern that most closely matches the URL.
	mux := http.NewServeMux()

	// "index.html" is served from "/" by convention
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(root)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8080", mux)
}

