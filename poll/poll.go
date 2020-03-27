package Poll

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

func (p *Poll) VoteUp(int choice) bool {
	if choice >= len(p.CHOICES) {
		return false
	}
	p.CHOICES[choice]++
	total++
	return true
}

func (p *Poll) VoteDown(int choices) bool {
	if choices >= len(p.choices) {
		return false
	}
	p.choices[choices]--
	total--
}
