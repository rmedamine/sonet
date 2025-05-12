package config

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var ErrExpiredSession = errors.New("session is expired")

type Session struct {
	Email     string
	Token     string
	UserId    int64
	ExpiresAt time.Time
}

type SessionManager struct {
	db *sql.DB
}

func NewSessionManager() {
	SESSION = &SessionManager{
		db: DB,
	}
}

func (s *SessionManager) CreateSession(email string, userId int64) (*Session, error) {
	// Delete existing sessions for the user
	if err := s.DeleteUserSessions(userId); err != nil {
		return nil, err
	}
	// Generate new token
	token, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	// Insert new session
	query := "INSERT INTO sessions (email, userId, token, expiresAt) VALUES (?, ?, ?, ?)"
	expTime := time.Now().Add(SESSION_EXP_TIME * time.Second)
	s.db.Exec(query, email, userId, token.String(), expTime)
	session, err := s.GetSession(token.String())
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionManager) GetSession(token string) (*Session, error) {
	query := `SELECT email, userId, expiresAt FROM sessions WHERE token = ?`
	var session Session

	row := s.db.QueryRow(query, token)
	err := row.Scan(&session.Email, &session.UserId, &session.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(session.Token)
		return nil, ErrExpiredSession
	}
	session.Token = token
	return &session, nil
}

func (s *SessionManager) DeleteSession(token string) error {
	query := "DELETE FROM sessions WHERE token = ?"
	_, err := s.db.Exec(query, token)
	return err
}

func (s *SessionManager) DeleteUserSessions(userId int64) error {
	query := "DELETE FROM sessions WHERE userId = ?"
	_, err := s.db.Exec(query, userId)
	return err
}
