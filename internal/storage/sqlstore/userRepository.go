package sqlstore

import (
	_ "github.com/mattn/go-sqlite3"
	"todolist/internal/storage"
)

type UserRepository struct {
	storage *Storage
}

func (r *UserRepository) Create(u *storage.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	r.storage.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", u.Username, u.EncryptedPassword)
	return r.storage.db.QueryRow("SELECT last_insert_rowid()").Scan(&u.ID)
}

func (r *UserRepository) FindByUsername(username string) (*storage.User, error) {
	u := &storage.User{}
	if err := r.storage.db.QueryRow("SELECT id, username, encrypted_password from users WHERE username = ?", username).Scan(&u.ID, &u.Username, &u.EncryptedPassword); err != nil {
		return nil, err
	}

	return u, nil
}
