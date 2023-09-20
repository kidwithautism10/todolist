package storage

import validation "github.com/go-ozzo/ozzo-validation"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Username, validation.Required, validation.Length(5, 100)),
		validation.Field(&u.Password, validation.Required, validation.Length(5, 100)),
	)
}
