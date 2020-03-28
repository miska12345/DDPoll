package poll

// Definition server poll struct
type Poll struct {
	Owner         string
	Accessibility bool // Private - 1 (members is effective) | Public - 0 (members is effective)
	Title         string
	Body          string
	Category      string
	Choices       []string // Description of each choices in the poill
	Members       []string
	Tags          []string
	Counts        []int64 // Vote counts, connected to choices by indices
	Total         int64
	ID            int64
}
