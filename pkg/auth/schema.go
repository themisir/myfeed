package auth

import (
	"github.com/labstack/echo/v4"
)

const ClaimsKey = "auth.claims"

type Claims interface {
	Id() string
}

type Schema interface {
	SignIn(c echo.Context, claims Claims) error
	SignOut(c echo.Context) error
	Authorize(c echo.Context) Claims
}

type claims struct {
	id string
}

func (c *claims) Id() string {
	return c.id
}
