package errors

var (
	ErrTeamExists   = NewDomainError(ErrCodeTeamExists, "team_name already exists")
	ErrPRExists     = NewDomainError(ErrCodePRExists, "PR id already exists")
	ErrPRMerged     = NewDomainError(ErrCodePRMerged, "cannot reassign on merged PR")
	ErrNotAssigned  = NewDomainError(ErrCodeNotAssigned, "reviewer is not assigned to this PR")
	ErrNoCandidate  = NewDomainError(ErrCodeNoCandidate, "no active replacement candidate in team")
	ErrNotFound     = NewDomainError(ErrCodeNotFound, "resource not found")
	ErrTeamNotFound = NewDomainError(ErrCodeNotFound, "team not found")
	ErrUserNotFound = NewDomainError(ErrCodeNotFound, "user not found")
	ErrPRNotFound   = NewDomainError(ErrCodeNotFound, "pull request not found")
)
