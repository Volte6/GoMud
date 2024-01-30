package usercommands

import (
	"github.com/volte6/mud/util"
)

func NewUserCommandResponse(uId int) util.MessageQueue {
	return util.NewMessageQueue(uId, 0)
}
