package storage

import validation "github.com/go-ozzo/ozzo-validation"

type Task struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Complete int    `json:"complete"`
	Date     string `json:"date"`
	Username string `json:"username"`
}

func (t *Task) ValidateTask() error {
	return validation.ValidateStruct(
		t,
		validation.Field(&t.Text, validation.Required, validation.Length(1, 100)),
	)
}
