package listing

type User interface {
	Id() string
	Email() string
	PasswordHash() string
}

type UserRepository interface {
	GetUserById(id string) (User, error)
	FindUserByEmail(email string) (User, error)
}
