package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
}

type User struct {
	ID           int
	Name         string
	Email        string
	HashPassword []byte
	Created      time.Time
}

type UserModel struct {
	DB      *pgxpool.Pool
	CONTEXT context.Context
}

func (m *UserModel) Insert(name, email, password string) error {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES(@name, @email, @hashed_password, NOW());`
	args := pgx.NamedArgs{
		"name":            name,
		"email":           email,
		"hashed_password": string(HashPassword),
	}

	_, err = m.DB.Exec(m.CONTEXT, stmt, args)
	if err != nil {
		var SQLError *pgconn.PgError
		if errors.As(err, &SQLError) {
			if SQLError.Code == "23505" && strings.Contains(SQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var (
		id              int
		hashed_password []byte
	)

	stmt := `select id, hashed_password FROM users WHERE email = $1;`

	err := m.DB.QueryRow(m.CONTEXT, stmt, email).Scan(&id, &hashed_password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashed_password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exist bool
	stmt := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"

	err := m.DB.QueryRow(m.CONTEXT, stmt, id).Scan(&exist)
	return exist, err
}
