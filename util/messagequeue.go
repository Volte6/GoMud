package util

type MessageType bool // only two types, so just use a bool

const (
	MsgUser MessageType = true
	MsgRoom MessageType = false
)

type newCommand struct {
	UserId        int
	MobInstanceId int
	Command       string
}

type message struct {
	MsgType        MessageType
	RoomId         int
	UserId         int
	Msg            string
	ExcludeUserIds []int // Optional additional userId's who should not receive a room message
}

// Response expected from user commands
type MessageQueue struct {
	Handled      bool         // Was this command recognized and handled? If not, display help etc.
	NextCommand  string       // Force another command to run right after this one
	UserId       int          // User ID
	MobId        int          // Mob ID
	CommandQueue []newCommand // new commands to queue up for any users/mobs
	userMsgCt    int
	roomMsgCt    int
}

func NewMessageQueue(uId int, mId int) MessageQueue {
	mq := MessageQueue{
		UserId:       uId,
		MobId:        mId,
		CommandQueue: make([]newCommand, 0),
	}

	return mq
}

func (u *MessageQueue) Len(lenType MessageType) int {
	if lenType == MsgUser {
		return u.userMsgCt
	}
	return u.roomMsgCt
}
