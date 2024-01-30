package spells

import "github.com/volte6/mud/util"

func Summon(sourceUserId int, sourceMobId int, details any, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := util.NewMessageQueue(sourceUserId, sourceMobId)

	// rest contains any special details of the spellcast, such as what creature to summon

	return response, nil
}
