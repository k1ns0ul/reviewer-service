package models

import "time"

type User struct {
	ID        string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// создание нового пользователя с валидацией
func NewUser(id, username, teamName string, isActive bool) (*User, error) {
	if id == "" {
		return nil, ErrInvalidUserID
	}
	if username == "" {
		return nil, ErrInvalidUsername
	}
	if teamName == "" {
		return nil, ErrInvalidTeamName
	}

	now := time.Now()
	return &User{
		ID:        id,
		Username:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// обновление статса активности
func (u *User) SetActive(isActive bool) {
	u.IsActive = isActive
	u.UpdatedAt = time.Now()
}

var (
	ErrInvalidUserID   = &ValidationError{Field: "user_id", Message: "cannot be empty"}
	ErrInvalidUsername = &ValidationError{Field: "username", Message: "cannot be empty"}
	ErrInvalidTeamName = &ValidationError{Field: "team_name", Message: "cannot be empty"}
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
