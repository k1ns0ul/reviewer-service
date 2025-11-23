package dto

type CreateTeamRequest struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	Team TeamData `json:"team"`
}

type TeamData struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
