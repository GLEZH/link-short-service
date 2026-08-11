package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/golang-jwt/jwt/v4"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

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

	if !uuidV4Pattern.MatchString(userID) {
		t.Fatalf("userID = %q, want uuid v4", userID)
	}
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
			if !uuidV4Pattern.MatchString(userID) {
				t.Fatalf("userID = %q, want uuid v4", userID)
			}
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
