package rooms

import "time"

type Sign struct {
	VisibleUserId int       // What user can see it? If 0, then everyone can see it.
	DisplayText   string    // What text to display
	Expires       time.Time // When this sign expires.
}
