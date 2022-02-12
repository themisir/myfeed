package adding

type (
	UserData struct {
		Email        string
		PasswordHash string
	}
	User interface {
		Id() string
		Email() string
		PasswordHash() string
	}
	UserRepository interface {
		AddUser(data UserData) (User, error)
	}
)
