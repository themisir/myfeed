package memory

import (
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"log"
	"strconv"
	"strings"
)

type UserData struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type user struct {
	UserData
}

func (u *user) Id() string {
	return u.UserData.Id
}

func (u *user) Email() string {
	return u.UserData.Email
}

func (u *user) PasswordHash() string {
	return u.UserData.PasswordHash
}

func NewUserRepository() *userRepository {
	return &userRepository{
		users: []*user{},
	}
}

type userRepository struct {
	users       []*user
	persistence Persistence
}

func (r *userRepository) Persist(p Persistence) {
	r.persistence = p
	r.load()
}

func (r *userRepository) save() {
	if r.persistence != nil {
		if err := r.persistence.Save(&r.users); err != nil {
			log.Printf("Failed to save feeds: %s", err)
		}
	}
}

func (r *userRepository) load() {
	if r.persistence != nil {
		if err := r.persistence.Load(&r.users); err != nil {
			log.Printf("Failed to load feeds: %s", err)
		}
	}
}

func (r *userRepository) AddUser(data adding.UserData) (adding.User, error) {
	item := &user{UserData{
		Id:           strconv.Itoa(len(r.users)),
		Email:        strings.ToLower(data.Email),
		PasswordHash: data.PasswordHash,
	}}

	r.users = append(r.users, item)
	r.save()
	return item, nil
}

func (r *userRepository) GetUserById(id string) (listing.User, error) {
	for _, user := range r.users {
		if user.Id() == id {
			return user, nil
		}
	}

	return nil, listing.ErrNotFound
}

func (r *userRepository) FindUserByEmail(email string) (listing.User, error) {
	normalizedEmail := strings.ToLower(email)

	for _, user := range r.users {
		if user.Email() == normalizedEmail {
			return user, nil
		}
	}

	return nil, listing.ErrNotFound
}
