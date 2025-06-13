package models

import (
    "database/sql"
    "time"
    
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID           int       `json:"id" db:"id"`
    Username     string    `json:"username" db:"username"`
    Email        string    `json:"email" db:"email"`
    PasswordHash string    `json:"-" db:"password_hash"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserService struct {
    db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
    return &UserService{db: db}
}

func (s *UserService) CreateUser(username, email, password string) (*User, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &User{
        Username:     username,
        Email:        email,
        PasswordHash: string(hashedPassword),
    }

    query := `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
    
    err = s.db.QueryRow(query, user.Username, user.Email, user.PasswordHash).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    
    return user, err
}

func (s *UserService) CreateGoogleUser(name, email, googleID string) (*User, error) {
    user := &User{
        Username:     name,
        Email:        email,
        PasswordHash: "", // No password for OAuth users
    }

    query := `
        INSERT INTO users (username, email, password_hash, google_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `
    
    err := s.db.QueryRow(query, user.Username, user.Email, user.PasswordHash, googleID).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    
    return user, err
}

func (s *UserService) GetUserByEmail(email string) (*User, error) {
    user := &User{}
    query := `
        SELECT id, username, email, password_hash, created_at, updated_at
        FROM users WHERE email = $1
    `
    
    err := s.db.QueryRow(query, email).Scan(
        &user.ID, &user.Username, &user.Email, &user.PasswordHash,
        &user.CreatedAt, &user.UpdatedAt,
    )
    
    return user, err
}

func (s *UserService) ValidatePassword(user *User, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
    return err == nil
}