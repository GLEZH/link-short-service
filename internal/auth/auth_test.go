package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func TestManager_BuildAndParseToken(t *testing.T) {
	manager := NewManager("test-secret")

	token, err := manager.BuildToken("user-id")
	if err != nil {
		t.Fatalf("BuildToken() error = %v", err)
	}

	userID, err := manager.ParseUserID(token)
	if err != nil {
		t.Fatalf("ParseUserID() error = %v", err)
	}

	if userID != "user-id" {
		t.Errorf("userID = %q, want %q", userID, "user-id")
	}
}

func TestManager_ParseUserIDWithoutUserID(t *testing.T) {
	manager := NewManager("test-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
	})

	tokenString, err := token.SignedString(manager.secret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = manager.ParseUserID(tokenString)
	if !errors.Is(err, entity.ErrUserIDNotFound) {
		t.Fatalf("ParseUserID() error = %v, want %v", err, entity.ErrUserIDNotFound)
	}
}

func TestNewUserID(t *testing.T) {
	userID, err := newUserID()
	if err != nil {
		t.Fatalf("newUserID() error = %v", err)
	}

	assertUUIDV4(t, userID)
}

func TestManager_WithAuth(t *testing.T) {
	t.Run("sets cookie when missing", func(t *testing.T) {
		manager := NewManager("test-secret")
		handler := manager.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				t.Fatal("user id is missing")
			}
			if userID == "" {
				t.Fatal("user id is empty")
			}
			assertUUIDV4(t, userID)
		}))

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if len(recorder.Result().Cookies()) != 1 {
			t.Fatalf("cookies count = %d, want 1", len(recorder.Result().Cookies()))
		}
	})

	t.Run("keeps valid cookie", func(t *testing.T) {
		manager := NewManager("test-secret")
		token, err := manager.BuildToken("user-id")
		if err != nil {
			t.Fatalf("BuildToken() error = %v", err)
		}

		handler := manager.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				t.Fatal("user id is missing")
			}
			if userID != "user-id" {
				t.Fatalf("userID = %q, want %q", userID, "user-id")
			}
		}))

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(&http.Cookie{Name: CookieName, Value: token})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if len(recorder.Result().Cookies()) != 0 {
			t.Fatalf("cookies count = %d, want 0", len(recorder.Result().Cookies()))
		}
	})
}

func assertUUIDV4(t *testing.T, value string) {
	t.Helper()

	userID, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("uuid.Parse(%q) error = %v", value, err)
	}
	if userID.Version() != 4 {
		t.Fatalf("uuid version = %d, want 4", userID.Version())
	}
}
