package main

import (
	"fmt"
	"net/http"
	"time"
	)


func main() {
	fmt.Println("test")
	mux := http.NewServeMux()
	
	srv := http.Server{
		Addr: ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Handler: mux,
	}

	srv.ListenAndServe()
}

