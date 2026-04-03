package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	conf "GanzamApi/conf"
	"GanzamApi/models"

	_ "github.com/denisenkom/go-mssqldb"
)

var ErrUserNotFound = errors.New("user not found")

type UserStore interface {
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	Create(ctx context.Context, req models.RegisterRequest, passwordHash string) (*models.User, error)
}

type MSSQLUserStore struct {
	db      *sql.DB
	columns map[string]bool
}

func NewMSSQLUserStore() (*MSSQLUserStore, error) {
	db, err := sql.Open("sqlserver", conf.GetDBConnectionString())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	columns, err := loadUsersColumns(ctx, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	required := []string{"Id", "PasswordHash"}
	for _, col := range required {
		if !columns[strings.ToLower(col)] {
			_ = db.Close()
			return nil, fmt.Errorf("Users table is missing required column %q. Available columns: %s", col, joinColumnNames(columns))
		}
	}

	return &MSSQLUserStore{db: db, columns: columns}, nil
}

func (s *MSSQLUserStore) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	if !s.columns["phone"] {
		return nil, fmt.Errorf("Users table is missing required column %q. Available columns: %s", "Phone", joinColumnNames(s.columns))
	}

	query := `
SELECT Id, Phone, Email, PasswordHash, FirstName, LastName, CreatedAt, UpdatedAt
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
		&user.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.IsActive = true
	user.Role = "customer"
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
	if !s.columns["phone"] {
		return nil, fmt.Errorf("Users table is missing required column %q. Available columns: %s", "Phone", joinColumnNames(s.columns))
	}

	query := `
INSERT INTO Users (Phone, Email, PasswordHash, FirstName, LastName)
OUTPUT INSERTED.Id, INSERTED.Phone, INSERTED.Email, INSERTED.FirstName, INSERTED.LastName, INSERTED.CreatedAt, INSERTED.UpdatedAt
VALUES (@p1, @p2, @p3, @p4, @p5)`

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
		&user.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	user.PasswordHash = passwordHash
	user.IsActive = true
	user.Role = "customer"
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

func loadUsersColumns(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `
SELECT COLUMN_NAME
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = 'Users'`)
	if err != nil {
		return nil, fmt.Errorf("load Users columns: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, fmt.Errorf("scan Users columns: %w", err)
		}
		columns[strings.ToLower(column)] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Users columns: %w", err)
	}

	if len(columns) == 0 {
		return nil, errors.New("Users table was not found in the current database")
	}

	return columns, nil
}

func joinColumnNames(columns map[string]bool) string {
	names := make([]string, 0, len(columns))
	for name := range columns {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
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
