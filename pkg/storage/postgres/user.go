package postgres

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"strings"
)

type userRepository struct {
	c                   *Connection
	addUserStmt         *sql.Stmt
	getUserByIdStmt     *sql.Stmt
	findUserByEmailStmt *sql.Stmt
}

const (
	addUserQuery         = `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	getUserByIdQuery     = `SELECT id, email, password_hash FROM users WHERE id = $1`
	findUserByEmailQuery = `SELECT id, email, password_hash FROM users WHERE email = $1`
)

func newUserRepository(c *Connection) (r *userRepository, err error) {
	r = &userRepository{c: c}
	err = c.Batch().
		Prepare(addUserQuery, &r.addUserStmt).
		Prepare(getUserByIdQuery, &r.getUserByIdStmt).
		Prepare(findUserByEmailQuery, &r.findUserByEmailStmt).
		Exec()
	return
}

func (r *userRepository) AddUser(data adding.UserData) (adding.User, error) {
	data.Email = strings.ToLower(data.Email)

	id := uuid.New().String()
	_, err := r.addUserStmt.Exec(id, data.Email, data.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user{
		id:           id,
		email:        data.Email,
		passwordHash: data.PasswordHash,
	}, nil
}

func (r *userRepository) GetUserById(id string) (listing.User, error) {
	var u user
	err := r.getUserByIdStmt.QueryRow(id).Scan(&u.id, &u.email, &u.passwordHash)
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (r *userRepository) FindUserByEmail(email string) (listing.User, error) {
	var u user
	err := r.findUserByEmailStmt.QueryRow(email).Scan(&u.id, &u.email, &u.passwordHash)
	if err != nil {
		return nil, err
	}
	return &u, err
}

type user struct {
	id           string
	email        string
	passwordHash string
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