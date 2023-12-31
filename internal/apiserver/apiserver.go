package apiserver

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"net/http"
	"todolist/internal/config"
	"todolist/internal/storage/sqlstore"
)

func Start() error {
	cfg := config.MustLoad()
	db, err := newDB(cfg.StoragePath)
	if err != nil {
		return err
	}

	defer db.Close()
	store := sqlstore.New(db)
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionsKey))
	srv := newServer(store, sessionStore)

	return http.ListenAndServe(cfg.Address, srv)
}

func newDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
