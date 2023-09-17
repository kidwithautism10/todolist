package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"todolist/internal/config"
)

func main() {
	cfg := config.MustLoad()

	router := chi.NewRouter()

	fs := http.FileServer(http.Dir("views/static/"))
	router.Handle("/static/*", http.StripPrefix("/static", fs))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("failed to start server")
	}

	log.Fatal("server stopped")
}
