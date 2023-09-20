package storage

type UserRepository struct {
	storage *Storage
}

func (r *UserRepository) Create(u *User) (*User, error) {
	if err := r.storage.db.QueryRow("INSERT INTO users (username, password) VALUES ($1, $2)", u.Username, u.Password); err != nil {
		return nil, err.Err()
	}
	r.storage.db.QueryRow("SELECT last_insert_rowid()").Scan(&u.ID)

	return u, nil
}

func (r *UserRepository) FindByUsername(username string) (*User, error) {
	return nil, nil
}
