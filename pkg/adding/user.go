package adding

type (
	UserData struct {
		Email        string
		Username     string
		PasswordHash string
	}
	User interface {
		Id() string
		Email() string
		Username() string
		PasswordHash() string
	}
	UserRepository interface {
		AddUser(data UserData) (User, error)
	}
)
