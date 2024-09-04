package conversations

var (
	conversations = make(map[int]Conversation)
)

type Conversation struct {
	Id             int
	MobInstanceId1 int
	MobInstanceId2 int
	Position       int
	ActionList     [][]string
}

func (c *Conversation) NextActions() []string {
	pos := c.Position
	c.Position++
	if c.Position >= len(c.ActionList) {
		return []string{}
	}
	
	return c.ActionList[pos]
}


func GetConversation(id int) *Conversation {

	if conversation, ok := conversations[id]; ok {

		if conversation.
		return &conversation
	}

	return nil
}
