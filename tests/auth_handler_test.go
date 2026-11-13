package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/middleware"
	authhandler "vendorsaathi/internal/modules/auth/handler"
	authmodel "vendorsaathi/internal/modules/auth/model"
	authservice "vendorsaathi/internal/modules/auth/service"
	"vendorsaathi/internal/platform/auth"
)

func setupAuthRouter(t *testing.T) (*gin.Engine, *authservice.AuthService, *auth.JWTManager, *mockUserRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	svc, userRepo, _, _, jwtMgr, _ := setupAuthTestEnv()
	h := authhandler.NewAuthHandler(svc)

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.POST("/api/v1/auth/login", h.Login)
	r.GET("/api/v1/me", middleware.AuthMiddleware(jwtMgr), h.Me)
	return r, svc, jwtMgr, userRepo
}

func postJSON(t *testing.T, r http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeErrorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v (body: %s)", err, w.Body.String())
	}
	return body.Error.Code
}

func TestLoginHandler_MalformedJSON(t *testing.T) {
	r, _, _, _ := setupAuthRouter(t)

	w := postJSON(t, r, "/api/v1/auth/login", "{")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
	if code := decodeErrorCode(t, w); code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", code)
	}
}

func TestLoginHandler_MissingFields(t *testing.T) {
	r, _, _, _ := setupAuthRouter(t)

	w := postJSON(t, r, "/api/v1/auth/login", `{}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
	if code := decodeErrorCode(t, w); code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", code)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	r, svc, _, _ := setupAuthRouter(t)

	phone := "+919900001111"
	password := "CorrectPassword123!"
	if _, err := svc.Register(context.Background(), authmodel.RegisterInput{
		Phone:    phone,
		Password: password,
	}); err != nil {
		t.Fatalf("register test user: %v", err)
	}

	t.Run("unknown phone", func(t *testing.T) {
		w := postJSON(t, r, "/api/v1/auth/login", `{"phone":"+910000000000","password":"CorrectPassword123!"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d (body: %s)", w.Code, w.Body.String())
		}
		if code := decodeErrorCode(t, w); code != "UNAUTHORIZED" {
			t.Fatalf("expected UNAUTHORIZED, got %q", code)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		w := postJSON(t, r, "/api/v1/auth/login", `{"phone":"`+phone+`","password":"WrongPassword123!"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d (body: %s)", w.Code, w.Body.String())
		}
		if code := decodeErrorCode(t, w); code != "UNAUTHORIZED" {
			t.Fatalf("expected UNAUTHORIZED, got %q", code)
		}
	})
}

func TestLoginHandler_SuccessAndProtectedRoute(t *testing.T) {
	r, svc, _, _ := setupAuthRouter(t)

	phone := "+919900002222"
	password := "CorrectPassword123!"
	if _, err := svc.Register(context.Background(), authmodel.RegisterInput{
		Phone:    phone,
		Password: password,
	}); err != nil {
		t.Fatalf("register test user: %v", err)
	}

	body, err := json.Marshal(authmodel.LoginInput{Phone: phone, Password: password})
	if err != nil {
		t.Fatalf("marshal login input: %v", err)
	}

	w := postJSON(t, r, "/api/v1/auth/login", string(body))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var out authmodel.TokenOutput
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode token output: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}
	if out.User.Phone != phone {
		t.Fatalf("expected user phone %q, got %q", phone, out.User.Phone)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+out.AccessToken)
	meW := httptest.NewRecorder()
	r.ServeHTTP(meW, req)
	if meW.Code != http.StatusOK {
		t.Fatalf("expected /me 200, got %d (body: %s)", meW.Code, meW.Body.String())
	}

	var me authmodel.MeOutput
	if err := json.Unmarshal(meW.Body.Bytes(), &me); err != nil {
		t.Fatalf("decode me output: %v", err)
	}
	if me.User.Phone != phone {
		t.Fatalf("expected /me user phone %q, got %q", phone, me.User.Phone)
	}
}

func TestAuthMiddleware_RejectsMissingMalformedAndRefreshTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtMgr := auth.NewJWTManager(
		"test-secret-key-at-least-32-bytes-long!",
		"vendorsaathi",
		"vendorsaathi-test",
		15*time.Minute,
		24*time.Hour,
	)

	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(jwtMgr), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetString("user_id")})
	})

	pair, err := jwtMgr.GenerateTokenPair("user-123", "session-1", []string{"ADMIN"}, []string{"admin:users"})
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{name: "missing header", header: "", want: http.StatusUnauthorized},
		{name: "malformed header", header: "Token abc", want: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer not-a-jwt", want: http.StatusUnauthorized},
		{name: "refresh token used as access token", header: "Bearer " + pair.RefreshToken, want: http.StatusUnauthorized},
		{name: "valid access token", header: "Bearer " + pair.AccessToken, want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("expected %d, got %d (body: %s)", tt.want, w.Code, w.Body.String())
			}
			if tt.want == http.StatusOK && !bytes.Contains(w.Body.Bytes(), []byte("user-123")) {
				t.Fatalf("expected user id in response, got %s", w.Body.String())
			}
		})
	}
}

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	jwtMgr := auth.NewJWTManager(
		"test-secret-key-at-least-32-bytes-long!",
		"vendorsaathi",
		"vendorsaathi-test",
		15*time.Minute,
		24*time.Hour,
	)

	pair, err := jwtMgr.GenerateTokenPair("user-456", "session-2", []string{"SUPER_ADMIN"}, []string{"admin:users"})
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}

	accessClaims, err := jwtMgr.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("validate access token: %v", err)
	}
	if accessClaims.UserID != "user-456" || accessClaims.SessionID != "session-2" {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}
	if accessClaims.TokenType != auth.TokenTypeAccess {
		t.Fatalf("expected access token type, got %q", accessClaims.TokenType)
	}
	if len(accessClaims.Roles) != 1 || accessClaims.Roles[0] != "SUPER_ADMIN" {
		t.Fatalf("unexpected roles: %v", accessClaims.Roles)
	}

	refreshClaims, err := jwtMgr.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("validate refresh token: %v", err)
	}
	if refreshClaims.TokenType != auth.TokenTypeRefresh {
		t.Fatalf("expected refresh token type, got %q", refreshClaims.TokenType)
	}

	if _, err := jwtMgr.ValidateAccessToken(pair.RefreshToken); !errors.Is(err, auth.ErrInvalidClaims) {
		t.Fatalf("expected refresh token to be rejected as access token, got %v", err)
	}
	if _, err := jwtMgr.ValidateRefreshToken(pair.AccessToken); !errors.Is(err, auth.ErrInvalidClaims) {
		t.Fatalf("expected access token to be rejected as refresh token, got %v", err)
	}
	if _, err := jwtMgr.ValidateAccessToken("not-a-jwt"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestLoginHandler_InternalErrorIsNotLeaked(t *testing.T) {
	r, _, _, userRepo := setupAuthRouter(t)
	userRepo.getUserByPhoneErr = errors.New(`ERROR: relation "users" does not exist (SQLSTATE 42P01)`)

	w := postJSON(t, r, "/api/v1/auth/login", `{"phone":"+919900003333","password":"CorrectPassword123!"}`)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d (body: %s)", w.Code, w.Body.String())
	}
	if code := decodeErrorCode(t, w); code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", code)
	}

	body := strings.ToLower(w.Body.String())
	for _, leaked := range []string{"relation", "sqlstate", "42p01", "does not exist"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("response leaked internal database detail %q: %s", leaked, w.Body.String())
		}
	}
}

// TestAdminProductionApiRouting ensures the deployed admin dashboard targets
// the production API for both application and health endpoints.
func TestAdminProductionApiRouting(t *testing.T) {
	raw, err := os.ReadFile("../admin/vercel.json")
	if err != nil {
		t.Fatalf("read admin/vercel.json: %v", err)
	}

	var cfg struct {
		Rewrites []struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
		} `json:"rewrites"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse admin/vercel.json: %v", err)
	}

	want := map[string]string{
		"/api/:path*":    "https://vendorsathi.onrender.com/api/:path*",
		"/health/:path*": "https://vendorsathi.onrender.com/health/:path*",
	}
	got := make(map[string]string, len(cfg.Rewrites))
	for _, rw := range cfg.Rewrites {
		got[rw.Source] = rw.Destination
	}
	for source, destination := range want {
		if got[source] != destination {
			t.Fatalf("rewrite %q = %q, want %q", source, got[source], destination)
		}
	}
}
