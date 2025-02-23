package templates

// A common structure used in templating
type NameDescription struct {
	Id          any  // optional identifier.
	Marked      bool // mark in some way?
	Name        string
	Description string
}
