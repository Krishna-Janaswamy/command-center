package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/service-virtualization-go/internal/auth"
	"github.com/example/service-virtualization-go/internal/middleware"
)

func TestRequireRoles_AdminAccess(t *testing.T) {
	jwtSvc := auth.NewJwtService("secret")

	// Generate token for admin user with "Admin" role
	adminToken, err := jwtSvc.GenerateToken("admin", "Admin", "AdminGroup")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := middleware.RequireRoles(jwtSvc, "Dev Ops")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for admin user, got %d", rec.Code)
	}
}

func TestRequireRoles_DevOpsAccess(t *testing.T) {
	jwtSvc := auth.NewJwtService("secret")

	devOpsToken, err := jwtSvc.GenerateToken("john", "Dev Ops", "QED_DEV_OPS")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := middleware.RequireRoles(jwtSvc, "Dev Ops")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+devOpsToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for Dev Ops user, got %d", rec.Code)
	}
}

func TestRequireRoles_ForbiddenUser(t *testing.T) {
	jwtSvc := auth.NewJwtService("secret")

	normalToken, err := jwtSvc.GenerateToken("user1", "Default User", "Users")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := middleware.RequireRoles(jwtSvc, "Dev Ops")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+normalToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for default user, got %d", rec.Code)
	}
}
