package models

const STATUS_ERROR = -1
const STATUS_OK = 1

const SEARCH_MAX_RESULT = 20

type Poll struct {
	id           int64
	host         string
	members      []string
	title        string
	content      string
	accessbility int8     // Private - 1 (members is effective) | Public - 0 (members is effective)
	choices      []string // Description of each choices in the poill
	counts       []int64  // Vote counts, connected to choices by indices
	total        int64
}
