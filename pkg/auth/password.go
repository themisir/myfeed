package auth

import "golang.org/x/crypto/bcrypt"

type PasswordHasher interface {
	// HashPassword generates hash from given password
	HashPassword(pwd string) string

	// CheckPasswordHash returns whether hash does matches hashed pwd value
	CheckPasswordHash(pwd string, hash string) bool
}

type bcryptHasher struct {
	cost int
}

func (b *bcryptHasher) HashPassword(pwd string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(pwd), b.cost)
	return string(bytes)
}

func (b *bcryptHasher) CheckPasswordHash(pwd string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
