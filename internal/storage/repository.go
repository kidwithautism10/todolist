package storage

type UserRepository interface {
	Create(*User) error
	FindByUsername(string) (*User, error)
}
