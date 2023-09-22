package sqlstore

import (
	"todolist/internal/storage"
)

type TaskRepository struct {
	storage *Storage
}

func (r *TaskRepository) CreateTask(t *storage.Task) error {
	if err := t.ValidateTask(); err != nil {
		return err
	}

	_, err := r.storage.db.Exec("INSERT INTO tasks (text, complete, date, user) VALUES (?, ?, ?, ?)", t.Text, 0, t.Date, t.Username)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskRepository) CompleteTask(id int) error {
	var complete int
	err := r.storage.db.QueryRow("SELECT complete from tasks WHERE id = ?", id).Scan(&complete)
	if err != nil {
		return err
	}
	if complete == 0 {
		_, err := r.storage.db.Exec("UPDATE tasks SET complete = 1 WHERE id = ?", id)
		if err != nil {
			return err
		}
	} else {
		_, err := r.storage.db.Exec("UPDATE tasks SET complete = 0 WHERE id = ?", id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskRepository) DeleteTask(id int) error {
	_, err := r.storage.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskRepository) RenderTask(username string) ([]storage.Task, error) {
	t := storage.Task{}
	ts := []storage.Task{}
	rows, err := r.storage.db.Query("SELECT * from tasks WHERE user = ?", username)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Text, &t.Complete, &t.Date, &t.Username)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	if len(ts) == 0 {
		t.Text = "здесь пока пусто :("
		ts = append(ts, t)
	}

	return ts, nil
}
