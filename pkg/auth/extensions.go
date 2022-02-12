package auth

import (
	"errors"
	"github.com/labstack/echo/v4"
)

var ErrNoHandler = errors.New("no handler found")

// GetHandler returns Handler linked to the context
func GetHandler(c echo.Context) (*Handler, error) {
	if handler, ok := c.Get(HandlerKey).(*Handler); ok {
		return handler, nil
	} else {
		return nil, ErrNoHandler
	}
}

// GetUserId returns authenticated user id from context
func GetUserId(c echo.Context) (string, error) {
	handler, err := GetHandler(c)
	if err != nil {
		return "", err
	} else {
		return handler.GetUserId(c), nil
	}
}
