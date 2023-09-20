package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"todolist/internal/config"
	"todolist/internal/storage"
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

	db, err := storage.New(cfg.StoragePath)
	if err != nil {
		fmt.Errorf("failed to init storage: %w", err)
		os.Exit(1)
	}

	_ = db

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("failed to start server")
	}

	log.Fatal("server stopped")
}
