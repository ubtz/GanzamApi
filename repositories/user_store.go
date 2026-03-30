package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"GanzamApi/config"
	"GanzamApi/models"

	_ "github.com/denisenkom/go-mssqldb"
)

var ErrUserNotFound = errors.New("user not found")

type UserStore interface {
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	Create(ctx context.Context, req models.RegisterRequest, passwordHash string) (*models.User, error)
}

type MSSQLUserStore struct {
	db *sql.DB
}

func NewMSSQLUserStore() (*MSSQLUserStore, error) {
	db, err := sql.Open("sqlserver", config.GetDBConnectionString())
	if err != nil {
		return nil, err
	}
	return &MSSQLUserStore{db: db}, nil
}

func (s *MSSQLUserStore) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
SELECT Id, Phone, Email, PasswordHash, FirstName, LastName, Role, IsActive, CreatedAt, UpdatedAt
FROM Users
WHERE Phone = @p1`

	var user models.User
	var email, firstName, lastName sql.NullString
	var updatedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.Phone,
		&email,
		&user.PasswordHash,
		&firstName,
		&lastName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if email.Valid {
		user.Email = &email.String
	}
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

func (s *MSSQLUserStore) Create(ctx context.Context, req models.RegisterRequest, passwordHash string) (*models.User, error) {
	query := `
INSERT INTO Users (Phone, Email, PasswordHash, FirstName, LastName, Role, IsActive)
OUTPUT INSERTED.Id, INSERTED.Phone, INSERTED.Email, INSERTED.FirstName, INSERTED.LastName, INSERTED.Role, INSERTED.IsActive, INSERTED.CreatedAt, INSERTED.UpdatedAt
VALUES (@p1, @p2, @p3, @p4, @p5, 'customer', 1)`

	var user models.User
	var email, firstName, lastName sql.NullString
	var updatedAt sql.NullTime

	err := s.db.QueryRowContext(
		ctx,
		query,
		req.Phone,
		nullableString(req.Email),
		passwordHash,
		nullableString(req.FirstName),
		nullableString(req.LastName),
	).Scan(
		&user.ID,
		&user.Phone,
		&email,
		&firstName,
		&lastName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	user.PasswordHash = passwordHash
	if email.Valid {
		user.Email = &email.String
	}
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

func nullableString(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

type MemoryUserStore struct {
	nextID int64
	users  map[string]*models.User
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		nextID: 1,
		users:  make(map[string]*models.User),
	}
}

func (s *MemoryUserStore) FindByPhone(_ context.Context, phone string) (*models.User, error) {
	user, ok := s.users[phone]
	if !ok {
		return nil, ErrUserNotFound
	}
	copyUser := *user
	return &copyUser, nil
}

func (s *MemoryUserStore) Create(_ context.Context, req models.RegisterRequest, passwordHash string) (*models.User, error) {
	if _, exists := s.users[req.Phone]; exists {
		return nil, fmt.Errorf("duplicate user")
	}

	now := time.Now()
	user := &models.User{
		ID:           s.nextID,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Role:         "customer",
		IsActive:     true,
		CreatedAt:    now,
	}
	s.nextID++

	if strings.TrimSpace(req.Email) != "" {
		email := strings.TrimSpace(req.Email)
		user.Email = &email
	}
	if strings.TrimSpace(req.FirstName) != "" {
		firstName := strings.TrimSpace(req.FirstName)
		user.FirstName = &firstName
	}
	if strings.TrimSpace(req.LastName) != "" {
		lastName := strings.TrimSpace(req.LastName)
		user.LastName = &lastName
	}

	s.users[req.Phone] = user
	copyUser := *user
	return &copyUser, nil
}
