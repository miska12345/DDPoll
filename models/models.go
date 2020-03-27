package models

const STATUS_ERROR = -1
const STATUS_OK = 1

const SEARCH_MAX_RESULT = 20

// Definition server poll struct
type Poll struct {
	ID           int64
	OWNER        string
	MEMBERS      []string
	TITLE        string
	BODY         string
	CATOGORY     string
	TAGS         []string
	ACCESSIBLITY int8     // Private - 1 (members is effective) | Public - 0 (members is effective)
	CHOICES      []string // Description of each choices in the poill
	COUNTS       []int64  // Vote counts, connected to choices by indices
	TOTAL        int64
}
