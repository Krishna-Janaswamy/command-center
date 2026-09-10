package service

import (
	"errors"
	"strings"
	"sync"

	"github.com/example/service-virtualization-go/internal/model"
)

type UserService interface {
	RegisterUser(username, password, email, adGroup string) (*model.User, error)
	FindByUsername(username string) *model.User
	CheckPassword(password, hashed string) bool
}

type inMemoryUserService struct {
	mu    sync.RWMutex
	users map[string]*model.User
}

func NewInMemoryUserService() UserService {
	return &inMemoryUserService{users: make(map[string]*model.User)}
}

func (s *inMemoryUserService) RegisterUser(username, password, email, adGroup string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[username]; ok {
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
	user := &model.User{
		Username: username,
		Password: string(hash),
		Email:    email,
		Role:     role,
		AdGroup:  adGroup,
	}
	s.users[username] = user
	return user, nil
}

func (s *inMemoryUserService) FindByUsername(username string) *model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.users[username]; ok {
		return u
	}
	return nil
}

func (s *inMemoryUserService) CheckPassword(password, hashed string) bool {
	return checkPasswordPBKDF2(password, hashed)
}
