package rooms

import (
	"time"
)

// There is a magic portal of Chuckles, magic portal of Henry here!
// There is a magical hole in the east wall here!
type TemporaryRoomExit struct {
	RoomId  int       // Where does it lead to?
	Title   string    // Does this exist have a special title?
	UserId  int       // Who created it?
	Expires time.Time // When will it be auto-cleaned up?
}

type RoomExit struct {
	RoomId       int
	Secret       bool     `yaml:"secret,omitempty"`
	MapDirection string   `yaml:"mapdirection,omitempty"` // Optionaly indicate the direction of this exit for mapping purposes
	Lock         GameLock `yaml:"lock,omitempty"`         // 0 - no lock. greater than zero = difficulty to unlock.
}

func (re RoomExit) HasLock() bool {
	return re.Lock.Difficulty > 0
}
