package mobcommands

import (
	"github.com/volte6/mud/util"
)

func NewMobCommandResponse(mId int) util.MessageQueue {
	return util.NewMessageQueue(0, mId)
}
