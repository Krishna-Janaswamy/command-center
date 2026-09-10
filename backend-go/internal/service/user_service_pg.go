package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/service-virtualization-go/internal/model"
)

type pgUserService struct {
	pool *pgxpool.Pool
}

func NewPostgresUserService(pool *pgxpool.Pool) UserService {
	return &pgUserService{pool: pool}
}

func (s *pgUserService) RegisterUser(username, password, email, adGroup string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check exists
	var exists bool
	row := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username)
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	hash, err := hashPasswordPBKDF2(password)
	if err != nil {
		return nil, err
	}

	role := "Default User"
	if strings.EqualFold(adGroup, "QED_DEV_OPS") || strings.EqualFold(adGroup, "admin") || strings.EqualFold(username, "admin") {
		role = "Dev Ops"
	}

	_, err = s.pool.Exec(ctx, "INSERT INTO users (username, password, email, role, adGroup, createdAt) VALUES ($1,$2,$3,$4,$5,now())",
		username, hash, email, role, adGroup)
	if err != nil {
		return nil, err
	}

	return &model.User{Username: username, Password: string(hash), Email: email, Role: role, AdGroup: adGroup}, nil
}

func (s *pgUserService) FindByUsername(username string) *model.User {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u model.User
	err := s.pool.QueryRow(ctx, "SELECT username, password, email, role, adGroup FROM users WHERE username=$1", username).Scan(&u.Username, &u.Password, &u.Email, &u.Role, &u.AdGroup)
	if err != nil {
		return nil
	}
	return &u
}

func (s *pgUserService) CheckPassword(password, hashed string) bool {
	return checkPasswordPBKDF2(password, hashed)
}
