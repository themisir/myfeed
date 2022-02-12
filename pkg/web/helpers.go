package web

import (
	"github.com/labstack/echo/v4"
	"github.com/themisir/myfeed/pkg/auth"
	"net/http"
)

func GetUserId(c echo.Context) (string, error) {
	return auth.GetUserId(c)
}

func Authorize(redirectOnFailure bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			handler, err := auth.GetHandler(c)
			if err != nil {
				c.Logger().Errorf("Failed to get auth handler: %s", err)
				return echo.ErrInternalServerError
			}

			if !handler.Authorize(c) {
				if redirectOnFailure {
					return c.Redirect(http.StatusSeeOther, "/login")
				}
			}

			return next(c)
		}
	}
}
