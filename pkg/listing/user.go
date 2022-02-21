package listing

type (
	User interface {
		Id() string
		Email() string
		Username() string
		PasswordHash() string
	}
	UserRepository interface {
		GetUserById(id string) (User, error)
		FindUserByUsername(username string) (User, error)
	}
)
