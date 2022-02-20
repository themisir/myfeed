package auth

import (
	"errors"
	"github.com/labstack/echo/v4"
)

const HandlerKey = "auth.handler"

var ErrInvalidPassword = errors.New("invalid password")

type Handler struct {
	hasher PasswordHasher
	schema Schema
}

// New creates authentication handler
func New(schema Schema) *Handler {
	return &Handler{
		hasher: &bcryptHasher{14},
		schema: schema,
	}
}

// SignIn checks user password then authenticates given user in current context
func (h *Handler) SignIn(c echo.Context, user User, password string) error {
	if h.hasher.CheckPasswordHash(password, user.PasswordHash()) {
		return h.schema.SignIn(c, &claims{id: user.Id()})
	} else {
		return ErrInvalidPassword
	}
}

// SignInWithoutPassword authenticates user without password checking in current context
func (h *Handler) SignInWithoutPassword(c echo.Context, user User) error {
	return h.schema.SignIn(c, &claims{id: user.Id()})
}

// GetUserId returns ID of currently authenticated or empty string not authenticated
func (h *Handler) GetUserId(c echo.Context) string {
	if claims := h.schema.Authorize(c); claims != nil {
		return claims.Id()
	} else {
		return ""
	}
}

// Init injects handler to the context
func (h *Handler) Init(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set(HandlerKey, h)
		return next(c)
	}
}

// Authorize authorizes current http context using configured authorization schema
func (h *Handler) Authorize(c echo.Context) bool {
	return h.schema.Authorize(c) != nil
}

// HashPassword returns hashed password
func (h *Handler) HashPassword(password string) string {
	return h.hasher.HashPassword(password)
}

// SignOut removes stored authentication data from request context
func (h *Handler) SignOut(c echo.Context) bool {
	return h.schema.SignOut(c) == nil
}
