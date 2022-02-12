package auth

type User interface {
	Id() string
	PasswordHash() string
}
