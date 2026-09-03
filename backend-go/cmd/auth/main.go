package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
    "github.com/jackc/pgx/v5/pgxpool"

    auth "github.com/example/service-virtualization-go/internal/auth"
    authhandler "github.com/example/service-virtualization-go/internal/handlers"
    "github.com/example/service-virtualization-go/internal/service"
)

func main() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Println("JWT_SECRET not set; using change-this-secret for compatibility")
        secret = "change-this-secret"
    }

    var userSvc service.UserService
    dbUrl := os.Getenv("DATABASE_URL")
    if dbUrl != "" {
        pool, err := pgxpool.New(context.Background(), dbUrl)
        if err != nil {
            log.Fatalf("failed to connect to database: %v", err)
        }
        // do not defer close in Lambda main
        userSvc = service.NewPostgresUserService(pool)
    } else {
        userSvc = service.NewInMemoryUserService()
    }
    jwtSvc := auth.NewJwtService(secret)
    h := authhandler.NewAuthHandler(userSvc, jwtSvc)

    mux := http.NewServeMux()
    mux.HandleFunc("/api/auth/register", h.Register)
    mux.HandleFunc("/api/auth/login", h.Login)
    mux.HandleFunc("/api/auth/verify-token", h.VerifyToken)

    // example protected endpoint for Lambda
    adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok","message":"admin area"}`))
    })

    // wrap with RBAC middleware
    mux.Handle("/api/admin-only", requireRBACLambda(jwtSvc, adminHandler))

    adapter := httpadapter.New(mux)
    lambda.Start(adapter.ProxyWithContext)
}

func requireRBACLambda(jwtSvc *auth.JwtService, h http.Handler) http.Handler {
    return requireRolesAdapterLambda(jwtSvc, h)
}

func requireRolesAdapterLambda(jwtSvc *auth.JwtService, h http.Handler) http.Handler {
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
        if claims.Role != "Dev Ops" {
            http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
            return
        }
        h.ServeHTTP(w, r)
    })
}
