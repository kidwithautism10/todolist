package storage

type Task struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Complete int    `json:"complete"`
	Date     string `json:"date"`
	Username string `json:"username"`
}
