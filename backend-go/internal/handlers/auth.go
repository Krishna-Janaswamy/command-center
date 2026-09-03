package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/example/service-virtualization-go/internal/auth"
	"github.com/example/service-virtualization-go/internal/service"
)

type AuthHandler struct {
	users      service.UserService
	jwtService *auth.JwtService
}

func NewAuthHandler(u service.UserService, j *auth.JwtService) *AuthHandler {
	return &AuthHandler{users: u, jwtService: j}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("Register endpoint hit from %s", r.RemoteAddr)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		AdGroup  string `json:"adGroup"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.users.RegisterUser(req.Username, req.Password, req.Email, req.AdGroup)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	token, err := h.jwtService.GenerateToken(user.Username, user.Role, user.AdGroup)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
			"adGroup":  user.AdGroup,
			"email":    user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login endpoint hit from %s", r.RemoteAddr)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	log.Printf("login attempt username=%s from=%s", req.Username, r.RemoteAddr)

	user := h.users.FindByUsername(req.Username)
	if user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !h.users.CheckPassword(req.Password, user.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtService.GenerateToken(user.Username, user.Role, user.AdGroup)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("user logged in: %s", user.Username)

	resp := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
			"adGroup":  user.AdGroup,
			"email":    user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	log.Printf("VerifyToken endpoint hit from %s", r.RemoteAddr)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	// Expect "Bearer <token>"
	var tokenStr string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr = authHeader[7:]
	} else {
		http.Error(w, "invalid authorization header", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtService.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("token valid for: %s", claims.Subject)

	resp := map[string]interface{}{
		"sub":     claims.Subject,
		"role":    claims.Role,
		"adGroup": claims.AdGroup,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
