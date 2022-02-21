package postgres

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"strings"
)

type userRepository struct {
	c                      *Connection
	addUserStmt            *sql.Stmt
	getUserByIdStmt        *sql.Stmt
	findUserByUsernameStmt *sql.Stmt
}

const (
	addUserQuery            = `INSERT INTO users (id, email, username, normalized_username, password_hash) VALUES ($1, $2, $3, $4, $5)`
	getUserByIdQuery        = `SELECT id, email, username, normalized_username, password_hash FROM users WHERE id = $1`
	findUserByUsernameQuery = `SELECT id, email, username, normalized_username, password_hash FROM users WHERE normalized_username = $1`
)

func newUserRepository(c *Connection) (r *userRepository, err error) {
	r = &userRepository{c: c}
	err = c.Batch().
		Prepare(addUserQuery, &r.addUserStmt).
		Prepare(getUserByIdQuery, &r.getUserByIdStmt).
		Prepare(findUserByUsernameQuery, &r.findUserByUsernameStmt).
		Exec()
	return
}

func (r *userRepository) AddUser(data adding.UserData) (adding.User, error) {
	id := uuid.New().String()
	normalizedUsername := strings.ToUpper(data.Username)
	_, err := r.addUserStmt.Exec(id, data.Email, data.Username, normalizedUsername, data.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user{
		id:                 id,
		email:              data.Email,
		username:           data.Username,
		normalizedUsername: normalizedUsername,
		passwordHash:       data.PasswordHash,
	}, nil
}

func (r *userRepository) GetUserById(id string) (listing.User, error) {
	var u user
	err := r.getUserByIdStmt.QueryRow(id).Scan(&u.id, &u.email, &u.username, &u.normalizedUsername, &u.passwordHash)
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (r *userRepository) FindUserByUsername(username string) (listing.User, error) {
	var u user
	err := r.findUserByUsernameStmt.QueryRow(strings.ToUpper(username)).Scan(&u.id, &u.email, &u.username, &u.normalizedUsername, &u.passwordHash)
	if err != nil {
		return nil, err
	}
	return &u, err
}

type user struct {
	id                 string
	email              string
	username           string
	normalizedUsername string
	passwordHash       string
}

func (u *user) Id() string {
	return u.id
}

func (u *user) Email() string {
	return u.email
}

func (u *user) PasswordHash() string {
	return u.passwordHash
}

func (u *user) Username() string {
	return u.username
}
