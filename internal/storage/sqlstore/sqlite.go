package sqlstore

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"todolist/internal/storage"
)

type Storage struct {
	db             *sql.DB
	userRepository *UserRepository
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) User() storage.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{s}

	return s.userRepository
}
