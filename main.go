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
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	metrics_html := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`, cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metrics_html))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
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
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	fmt.Println("starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	fmt.Println(err)
}

