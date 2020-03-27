func (p *models.Poll) VoteUp(int choice) bool {
	if choice >= len(p.CHOICES) {
		return false
	}
	p.CHOICES[choice]++
	total++
	return true
}

func (p *models.Poll) VoteDown(int choices) bool {
	if choices >= len(p.choices) {
		return false
	}
	p.choices[choices]--
	total--
}
