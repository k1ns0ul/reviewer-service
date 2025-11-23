package models

import (
	"reviewer-service/internal/domain/enums"
	"time"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            enums.PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	MergedAt          *time.Time
}

// создание нового PR
func NewPullRequest(id, name, authorID string) (*PullRequest, error) {
	if id == "" {
		return nil, &ValidationError{Field: "pull_request_id", Message: "cannot be empty"}
	}
	if name == "" {
		return nil, &ValidationError{Field: "pull_request_name", Message: "cannot be empty"}
	}
	if authorID == "" {
		return nil, &ValidationError{Field: "author_id", Message: "cannot be empty"}
	}

	now := time.Now()
	return &PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            enums.StatusOpen,
		AssignedReviewers: make([]string, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// проверка на мерж PR
func (pr *PullRequest) IsMerged() bool {
	return pr.Status == enums.StatusMerged
}

// отметка PR как смерженный
func (pr *PullRequest) Merge() {
	if pr.IsMerged() {
		return
	}
	now := time.Now()
	pr.Status = enums.StatusMerged
	pr.MergedAt = &now
	pr.UpdatedAt = now
}

// назначение ревьюера
func (pr *PullRequest) AssignReviewers(reviewers []string) {
	pr.AssignedReviewers = reviewers
	pr.UpdatedAt = time.Now()
}

// проверка на назначенного ревьюера
func (pr *PullRequest) IsReviewerAssigned(userID string) bool {
	for _, r := range pr.AssignedReviewers {
		if r == userID {
			return true
		}
	}
	return false
}

// заменение ревьювера
func (pr *PullRequest) ReplaceReviewer(oldUserID, newUserID string) bool {
	for i, r := range pr.AssignedReviewers {
		if r == oldUserID {
			pr.AssignedReviewers[i] = newUserID
			pr.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}
