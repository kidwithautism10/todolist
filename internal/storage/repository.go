package storage

type UserRepository interface {
	Create(*User) error
	Find(int) (*User, error)
	FindByUsername(string) (*User, error)
}

type TaskRepository interface {
	CreateTask(*Task) error
	CompleteTask(int) error
	DeleteTask(int) error
	RenderTask(string) ([]Task, error)
}
