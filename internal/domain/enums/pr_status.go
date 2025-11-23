package enums

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

func (s PRStatus) IsValid() bool {
	switch s {
	case StatusOpen, StatusMerged:
		return true
	}
	return false
}

func (s PRStatus) String() string {
	return string(s)
}
