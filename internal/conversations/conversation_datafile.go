package conversations

type ConversationData struct {
	// A map of lowercase names of "Initiator" (#1) to array of
	// "Participant" (#2) names allowed to use this conversation.
	Supported map[string][]string `yaml:"Supported"`
	// A list of command lists, each prefixed with which Mob should execute
	// the action (#1 or #2)
	Conversation [][]string `yaml:"Conversation"`
}
