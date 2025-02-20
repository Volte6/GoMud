package events

// Used to apply or remove buffs
type Buff struct {
	UserId        int
	MobInstanceId int
	BuffId        int
}

func (b Buff) Type() string { return `Buff` }

// Used for giving/taking quest progress
type Quest struct {
	UserId     int
	QuestToken string
}

func (q Quest) Type() string { return `Quest` }

// For special room-targetting actions
type RoomAction struct {
	RoomId       int
	SourceUserId int
	SourceMobId  int
	Action       string
	Details      any
	WaitTurns    int
}

func (r RoomAction) Type() string { return `RoomAction` }

// Used for Input from players/mobs
type Input struct {
	UserId        int
	MobInstanceId int
	InputText     string
	WaitTurns     int
}

func (i Input) Type() string { return `Input` }

// Messages that are intended to reach all users on the system
type Broadcast struct {
	Text            string
	SkipLineRefresh bool
}

func (b Broadcast) Type() string { return `Broadcast` }

type Message struct {
	UserId          int
	ExcludeUserIds  []int
	RoomId          int
	Text            string
	IsQuiet         bool // whether it can only be heard by superior "hearing"
	IsCommunication bool // If true, this is a communication such as "say" or "emote"
}

func (m Message) Type() string { return `Message` }

// Messages that are intended to reach all users on the system
type WebClientCommand struct {
	ConnectionId uint64
	Text         string
}

func (b WebClientCommand) Type() string { return `WebClientCommand` }

// Messages that are intended to reach all users on the system
type GMCPIn struct {
	ConnectionId uint64
	Command      string
	Json         []byte
}

func (b GMCPIn) Type() string { return `GMCP` }

// Messages that are intended to reach all users on the system
type GMCPOut struct {
	ConnectionId uint64
	UserId       int
	Payload      any
}

func (b GMCPOut) Type() string { return `GMCP` }

// Messages that are intended to reach all users on the system
type System struct {
	Command string
}

func (s System) Type() string { return `System` }

// Messages that are intended to reach all users on the system
type MSP struct {
	UserId    int
	SoundType string // SOUND or MUSIC
	SoundFile string
	Volume    int    // 1-100
	Category  string // special category/type for MSP string
}

func (m MSP) Type() string { return `MSP` }
