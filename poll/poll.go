package poll

// Definition server poll struct
type Poll struct {
	Owner         string
	Accessibility int8 // Private - 1 (members is effective) | Public - 0 (members is effective)
	Title         string
	Body          string
	Category      string
	Members       []string
	Tags          []string
	Choices       []string // Description of each choices in the poill
	Counts        []int64  // Vote counts, connected to choices by indices
	Total         int64
	ID            int64
}
