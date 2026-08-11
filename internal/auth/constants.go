package auth

import "time"

const (
	CookieName = "auth_token"
	tokenTTL   = 24 * time.Hour * 365
)
