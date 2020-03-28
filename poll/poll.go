// package poll contain data structures for poll service
package poll

import (
	"time"
)

// Definition server poll struct
type Poll struct {
	PID        string
	Owner      string
	Public     bool // Private - False | Public - True
	Title      string
	Content    string
	Category   string
	Choices    []string // Description of each choices in the poill
	Tags       []string
	Votes      []uint64 // Vote counts, connected to choices by indices
	numVoted   uint64
	numViewed  uint64
	numStarred uint64
	CreateTime time.Time
	EndTime    time.Time
}
