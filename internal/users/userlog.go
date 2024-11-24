package users

import (
	"time"

	"github.com/volte6/gomud/internal/util"
)

const (
	LogMinAllocation = 50
	LogMaxAllocation = 1000
)

type UserLogEntry struct {
	Category  string    // optional category such as "combat" "communication" etc
	WhenRound uint64    // Game round it occured
	WhenTime  time.Time // Actual time it occured
	What      string    // String describing occurance
}

type UserLog []UserLogEntry

func (ul *UserLog) Add(cat string, message string) {

	if LogMinAllocation < 1 { // disables log if <1
		return
	}

	if len(*ul) == cap(*ul) {

		if cap(*ul) < LogMaxAllocation {

			newUL := make(UserLog, len(*ul), cap(*ul)+LogMinAllocation)
			copy(newUL, *ul)
			*ul = newUL

		} else {

			newUL := make(UserLog, len(*ul)-1, LogMaxAllocation)
			copy(newUL, (*ul)[1:])
			*ul = newUL

		}

	}

	newEntry := UserLogEntry{
		Category:  cat,
		WhenTime:  time.Now(),
		WhenRound: util.GetRoundCount(),
		What:      message,
	}

	*ul = append(*ul, newEntry)

}

func (ul *UserLog) Items(yield func(UserLogEntry) bool) {
	for _, v := range *ul {
		if !yield(v) {
			return
		}
	}
}
