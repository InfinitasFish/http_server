package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	)

type apiConfig struct {
	// atomic is thread-safe type for multiple goroutines
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// injecting some code before serving original Handler
		r.Header.Add("Cache-Control", "no-cache")
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetHits() {
	cfg.fileserverHits.Store(0)
}

func main() {
	root := http.Dir(".")
	fmt.Println(root)

	// ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request against a list 
	// of registered patterns and calls the handler for the pattern that most closely matches the URL.
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	// "index.html" is served from "/" by convention
	root_handler := http.StripPrefix("/app", http.FileServer(root))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(root_handler))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %v\n", apiCfg.fileserverHits.Load())))
	})

	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		apiCfg.resetHits()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Reset hits: %v\n", apiCfg.fileserverHits.Load())))
	})

	fmt.Println("starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	fmt.Println(err)
}

