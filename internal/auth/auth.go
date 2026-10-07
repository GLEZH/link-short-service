package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

// Claims contains the authenticated user ID
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// Manager creates and validates authentication tokens
type Manager struct {
	secret []byte
}

// NewManager creates an authentication manager
func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

// WithAuth adds user identity to each request context
func (m *Manager) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.userIDFromRequest(r)
		switch {
		case err == nil:
		case errors.Is(err, entity.ErrUserIDNotFound):
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			userID, err = m.setNewUserCookie(w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
	})
}

func (m *Manager) setNewUserCookie(w http.ResponseWriter) (string, error) {
	userID, err := newUserID()
	if err != nil {
		return "", err
	}
	if err = m.setCookie(w, userID); err != nil {
		return "", err
	}

	return userID, nil
}

func (m *Manager) userIDFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", err
	}

	return m.ParseUserID(cookie.Value)
}

func (m *Manager) setCookie(w http.ResponseWriter, userID string) error {
	token, err := m.BuildToken(userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(tokenTTL),
		HttpOnly: true,
	})
	return nil
}

// BuildToken creates a signed token for a user
func (m *Manager) BuildToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	})

	return token.SignedString(m.secret)
}

// ParseUserID validates a token and returns its user ID
func (m *Manager) ParseUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("token is not valid")
	}
	if claims.UserID == "" {
		return "", entity.NewUserIDNotFoundError()
	}

	return claims.UserID, nil
}

// WithUserID stores a user ID in a context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext returns the user ID stored in a context
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok && userID != ""
}

func newUserID() (string, error) {
	userID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("generate user id: %w", err)
	}

	return userID.String(), nil
}
