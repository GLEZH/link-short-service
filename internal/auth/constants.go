package auth

import "time"

const (
	CookieName = "auth_token"
	tokenTTL   = 365 * 24 * time.Hour
)
