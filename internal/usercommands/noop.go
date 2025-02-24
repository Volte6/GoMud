package usercommands

import (
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

// This is a no-op, does nothing
func Noop(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {
	return true, nil
}
