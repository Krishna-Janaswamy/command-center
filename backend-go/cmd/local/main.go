package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	auth "github.com/example/service-virtualization-go/internal/auth"
	authhandler "github.com/example/service-virtualization-go/internal/handlers"
	"github.com/example/service-virtualization-go/internal/service"
	"github.com/example/service-virtualization-go/internal/virtualization"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	addr := ":3001"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}

	var userSvc service.UserService
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl != "" {
		pool, err := pgxpool.New(context.Background(), dbUrl)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer pool.Close()
		userSvc = service.NewPostgresUserService(pool)
	} else {
		// try SQLite file (mimic Spring Boot fallback to file DB)
		sqlitePath := os.Getenv("SQLITE_PATH")
		if sqlitePath == "" {
			sqlitePath = "../backend/data/data.db"
		}
		s, err := service.NewSQLiteUserService(sqlitePath)
		if err != nil {
			log.Printf("failed to open sqlite at %s: %v, falling back to in-memory", sqlitePath, err)
			userSvc = service.NewInMemoryUserService()
		} else {
			userSvc = s
		}
	}
	jwtSvc := auth.NewJwtService(secret)
	handler := authhandler.NewAuthHandler(userSvc, jwtSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/register", handler.Register)
	mux.HandleFunc("/api/auth/login", handler.Login)
	mux.HandleFunc("/api/auth/verify-token", handler.VerifyToken)
	mux.HandleFunc("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	store := virtualization.NewConfiguredStore(context.Background())
	management := &virtualization.ManagementHandler{Store: store}
	runtime := &virtualization.RuntimeHandler{Store: store, Blobs: virtualization.NewConfiguredBlobStore(context.Background())}
	capture := &virtualization.CaptureHandler{Store: store, Blobs: virtualization.NewConfiguredBlobStore(context.Background())}
	health := &virtualization.HealthCheckHandler{}
	for _, pattern := range []string{"/api/registry", "/api/registry/", "/api/stubs", "/api/stubs/", "/api/requests", "/api/requests/", "/api/recorded-requests", "/api/settings", "/api/settings/", "/api/analytics", "/api/mock", "/api/mock/"} {
		mux.Handle(pattern, management)
	}
	mux.Handle("/api/capture", capture)
	mux.Handle("/api/health-check", health)
	mux.Handle("/", runtime)

	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"admin area"}`))
	})

	mux.Handle("/api/admin-only", requireRBAC(jwtSvc, adminHandler))

	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:8080"
	}

	handlerWithCors := cors(mux, frontendOrigin)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handlerWithCors,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

// cors and requireRBAC reused from previous main implementations
func cors(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireRBAC(jwtSvc *auth.JwtService, h http.Handler) http.Handler {
	return (&middlewareWrapper{jwt: jwtSvc}).Wrap(h)
}

type middlewareWrapper struct {
	jwt *auth.JwtService
}

func (m *middlewareWrapper) Wrap(h http.Handler) http.Handler {
	return requireRolesAdapter(m.jwt, h)
}

func requireRolesAdapter(jwtSvc *auth.JwtService, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		token := authHeader[7:]
		claims, err := jwtSvc.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		isAdmin := strings.EqualFold(claims.Subject, "admin") ||
			strings.EqualFold(claims.AdGroup, "admin") ||
			strings.EqualFold(claims.AdGroup, "QED_DEV_OPS") ||
			strings.EqualFold(claims.Role, "admin") ||
			strings.EqualFold(claims.Role, "Dev Ops")

		if !isAdmin {
			http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
