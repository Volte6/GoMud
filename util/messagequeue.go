package util

import (
	"errors"
	"strings"

	"github.com/volte6/mud/term"
)

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
	messages     []message    // Messages to send to the user
	CommandQueue []newCommand // new commands to queue up for any users/mobs
	userMsgCt    int
	roomMsgCt    int
}

func NewMessageQueue(uId int, mId int) MessageQueue {
	mq := MessageQueue{
		UserId:       uId,
		MobId:        mId,
		messages:     make([]message, 0, 2),
		CommandQueue: make([]newCommand, 0),
	}

	return mq
}

// Returns whether messages are waiting for dispatch.
func (u *MessageQueue) Pending() bool {
	return len(u.messages) > 0
}

func (u *MessageQueue) Len(lenType MessageType) int {
	if lenType == MsgUser {
		return u.userMsgCt
	}
	return u.roomMsgCt
}

func (u *MessageQueue) SendUserMessage(userId int, msg string) {
	msg += term.CRLFStr
	u.messages = append(u.messages, message{
		MsgType: MsgUser,
		UserId:  userId,
		Msg:     msg,
	})

	u.userMsgCt++
}

func (u *MessageQueue) SendUserMessages(userId int, msgs []string) {

	for _, msg := range msgs {
		msg += term.CRLFStr
		u.messages = append(u.messages, message{
			MsgType: MsgUser,
			UserId:  userId,
			Msg:     msg,
		})
	}

	u.userMsgCt += len(msgs)
}

// Gets all user messages to a specific user as a single string.
func (u *MessageQueue) GetUserMessagesAsString(userId int) string {
	userMessages := strings.Builder{}
	for _, message := range u.messages {
		if message.MsgType == MsgUser && message.UserId == userId {
			userMessages.WriteString(message.Msg)
		}
	}
	return userMessages.String()
}

func (u *MessageQueue) AbsorbMessages(response MessageQueue) {
	u.messages = append(u.messages, response.messages...)

	u.userMsgCt += response.userMsgCt
	u.roomMsgCt += response.roomMsgCt
}

func (u *MessageQueue) GetNextMessage() (message, error) {

	returnMsg := message{}

	if len(u.messages) == 0 {
		return returnMsg, errors.New("no messages")
	}

	returnMsg = u.messages[0]
	u.messages = u.messages[1:]

	if returnMsg.MsgType == MsgUser {
		u.userMsgCt--
	} else {
		u.roomMsgCt--
	}

	return returnMsg, nil
}
