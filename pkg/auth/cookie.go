package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type CookieOptions struct {
	Name     string
	Lifetime time.Duration
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

func (c *CookieOptions) Cookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(c.Lifetime),
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
		SameSite: c.SameSite,
	}
}

func (c *CookieOptions) Expired() *http.Cookie {
	cookie := c.Cookie("")
	cookie.Expires = time.Now()
	return cookie
}

func CookieSchema(secret []byte, lifetime time.Duration) *cookieSchema {
	return &cookieSchema{
		secret:   secret,
		lifetime: lifetime,

		Cookie: CookieOptions{
			Name:     "auth",
			Lifetime: lifetime,
			HttpOnly: true,
		},
	}
}

type cookieSchema struct {
	secret   []byte
	lifetime time.Duration

	Cookie   CookieOptions
	Issuer   string
	Audience string
}

func (s *cookieSchema) SignIn(c echo.Context, claims Claims) error {
	jwtClaims := &jwt.StandardClaims{
		Id:        uuid.New().String(),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(s.lifetime).Unix(),
		Subject:   claims.Id(),
		Issuer:    s.Issuer,
		Audience:  s.Audience,
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	// Generate signed and encoded token
	t, err := token.SignedString(s.secret)
	if err != nil {
		return err
	}

	// Set response cookie from generated token
	c.SetCookie(s.Cookie.Cookie(t))

	// Save claims to context
	c.Set(ClaimsKey, claims)

	return nil
}

func (s *cookieSchema) SignOut(c echo.Context) error {
	// Replace existing cookie with expired one
	c.SetCookie(s.Cookie.Expired())

	// Remove claims from context
	c.Set(ClaimsKey, nil)

	return nil
}

func (s *cookieSchema) Authorize(c echo.Context) Claims {
	// Check retrieving claims from context
	if claims, ok := c.Get(ClaimsKey).(Claims); ok {
		return claims
	}

	// Read cookie
	cookie, err := c.Cookie(s.Cookie.Name)
	if err != nil {
		return nil
	}

	// Decode and validate JWT token
	jwtClaims := new(jwt.StandardClaims)
	token, err := jwt.ParseWithClaims(cookie.Value, jwtClaims, s.keyFunc)
	if err != nil || !token.Valid {
		return nil
	}

	claims := &claims{jwtClaims.Subject}

	// Save claims to context
	c.Set(ClaimsKey, claims)

	return claims
}

func (s *cookieSchema) keyFunc(t *jwt.Token) (interface{}, error) {
	claims := t.Claims.(*jwt.StandardClaims)

	// Check JWT issuer
	if s.Issuer != "" && s.Issuer != claims.Issuer {
		return nil, fmt.Errorf("invalid jwt issuer")
	}

	// Check JWT audience
	if s.Audience != "" && s.Audience != claims.Audience {
		return nil, fmt.Errorf("invalid jwt audience")
	}

	return s.secret, nil
}
