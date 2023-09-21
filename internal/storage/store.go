package storage

type Store interface {
	User() UserRepository
	Task() TaskRepository
}
