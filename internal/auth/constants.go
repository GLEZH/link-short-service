package auth

import "time"

const (
	// CookieName is the authentication cookie name
	CookieName = "auth_token"
	tokenTTL   = 365 * 24 * time.Hour
)
