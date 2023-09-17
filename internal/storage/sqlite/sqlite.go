package sqlite

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage/sqlite/sqlite"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	stmt, err := db.Prepare("")
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
}
