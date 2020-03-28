// package poll contain data structures for poll service
package poll

import (
	"time"
)

// Definition server poll struct
type Poll struct {
	Owner         string
	Accessibility bool // Private - True | Public - 0
	Title         string
	Body          string
	Category      string
	Choices       []string // Description of each choices in the poill
	Tags          []string
	Counts        []int64 // Vote counts, connected to choices by indices
	Total         int64
	ID            int64
	CreateTime    time.Time
	EndTime       time.Time
}
