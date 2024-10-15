package users

import (
	"time"

	"github.com/volte6/mud/items"
)

type Inbox []Message

type Message struct {
	FromUserId int
	FromName   string
	Message    string
	Item       items.Item
	Gold       int
	Read       bool
	DateSent   time.Time
}

func (i *Inbox) Add(msg Message) {

	msg.DateSent = time.Now()

	newInbox := &Inbox{msg}

	if i == nil {
		(*i) = *newInbox
		return
	}

	(*i) = append(*newInbox, (*i)...)
}

func (i *Inbox) Empty() {
	(*i) = Inbox{}
}
