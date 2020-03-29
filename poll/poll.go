// package poll contain data structures for poll service
package poll

import (
	"time"
)

// MaxSearchResult is the maximum search result
const MaxSearchResult = 20

// MinOptions is the lower-bound for options
const MinOptions = 2

// MaxOptions is the upper-bound for options
const MaxOptions = 10

// RequiredPollElements is the number of elements required
const RequiredPollElements = 4

// CreateParamLength is the number of parameters required by rpc create call
const CreateParamLength = MinOptions + RequiredPollElements

// Definition server poll struct
type Poll struct {
	PID        string `json:"_id,omitempty" bson:"_id,omitempty"`
	Owner      string
	Public     bool // Private - False | Public - True
	Title      string
	Content    string
	Category   string
	Choices    []string // Description of each choices in the poill
	Tags       []string
	Votes      []uint64 // Vote counts, connected to choices by indices
	Stars      uint64
	NumVoted   uint64
	NumViewed  uint64
	NumStarred uint64
	CreateTime time.Time
	EndTime    time.Time
}
