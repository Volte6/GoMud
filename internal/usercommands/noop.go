package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// This is a no-op, does nothing
func Noop(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	return true, nil
}
