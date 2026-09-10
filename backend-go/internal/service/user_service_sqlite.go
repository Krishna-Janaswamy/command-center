package service

import (
    "database/sql"
    "errors"
    "strings"
    "time"

    _ "github.com/mattn/go-sqlite3"

    "github.com/example/service-virtualization-go/internal/model"
)

type sqliteUserService struct {
    db *sql.DB
}

func NewSQLiteUserService(dbPath string) (UserService, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(1)

    // ensure users table exists (mimic Spring DbService schema)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        username TEXT PRIMARY KEY,
        password TEXT NOT NULL,
        email TEXT,
        role TEXT,
        adGroup TEXT,
        createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        db.Close()
        return nil, err
    }

    return &sqliteUserService{db: db}, nil
}

func (s *sqliteUserService) RegisterUser(username, password, email, adGroup string) (*model.User, error) {
    // check exists
    var exists int
    err := s.db.QueryRow("SELECT COUNT(1) FROM users WHERE username = ?", username).Scan(&exists)
    if err != nil {
        return nil, err
    }
    if exists > 0 {
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

    _, err = s.db.Exec("INSERT INTO users (username, password, email, role, adGroup, createdAt) VALUES (?,?,?,?,?,datetime('now'))",
        username, hash, email, role, adGroup)
    if err != nil {
        return nil, err
    }
    return &model.User{Username: username, Password: string(hash), Email: email, Role: role, AdGroup: adGroup}, nil
}

func (s *sqliteUserService) FindByUsername(username string) *model.User {
    var u model.User
    err := s.db.QueryRow("SELECT username, password, email, role, adGroup FROM users WHERE username = ?", username).Scan(&u.Username, &u.Password, &u.Email, &u.Role, &u.AdGroup)
    if err != nil {
        return nil
    }
    return &u
}

func (s *sqliteUserService) CheckPassword(password, hashed string) bool {
    return checkPasswordPBKDF2(password, hashed)
}
