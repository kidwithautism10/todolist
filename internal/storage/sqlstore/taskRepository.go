package sqlstore

type TaskRepository struct {
	storage *Storage
}

func (r *TaskRepository) CreateTask(text string, date string, user string) error {
	_, err := r.storage.db.Exec("INSERT INTO tasks (text, complete, date, user) VALUES (?, ?, ?, ?)", text, 0, date, user)
	if err != nil {
		return err
	}
	return nil
}
