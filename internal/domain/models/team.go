package models

import "time"

type Team struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}

// создание новлй команды
func NewTeam(name string) (*Team, error) {
	if name == "" {
		return nil, &ValidationError{Field: "team_name", Message: "cannot be empty"}
	}

	now := time.Now()
	return &Team{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
