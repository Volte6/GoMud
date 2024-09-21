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
	UserId         int
	ExcludeUserIds []int
	RoomId         int
	Text           string
	IsQuiet        bool // whether it can only be heard by superior "hearing"
}

func (m Message) Type() string { return `Message` }

type ClientSettings struct {
	ConnectionId uint64
	ScreenWidth  uint32
	ScreenHeight uint32
	Monochrome   bool
}

func (c ClientSettings) Type() string { return `ClientSettings` }

// Messages that are intended to reach all users on the system
type WebClientCommand struct {
	ConnectionId uint64
	Text         string
}

func (b WebClientCommand) Type() string { return `WebClientCommand` }
