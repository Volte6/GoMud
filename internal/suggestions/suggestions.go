package suggestions

type Suggestions struct {
	suggestions []string
	pos         int
}

func (s *Suggestions) Count() int {
	return len(s.suggestions)
}

func (s *Suggestions) Clear() {
	s.suggestions = []string{}
	s.pos = 0
}

func (s *Suggestions) Set(suggestions []string) {
	s.suggestions = suggestions
	s.pos = 0
}

func (s *Suggestions) Next() string {
	if len(s.suggestions) < 1 {
		return ``
	}
	if s.pos >= len(s.suggestions) {
		s.pos = 0
	}
	s.pos++
	return s.suggestions[s.pos-1]
}
