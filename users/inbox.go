package users

import (
	"time"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
)

type Inbox []Message

type Message struct {
	FromUserId int
	FromName   string
	Message    string
	Item       *items.Item
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

func (i *Inbox) Count(unread ...bool) int {

	ct := 0
	for _, msg := range *i {
		if len(unread) > 0 && unread[0] {
			if !msg.Read {
				ct++
			}
		} else {
			ct++
		}
	}
	return ct
}

func (i *Inbox) CountRead() int {
	readCt := 0
	for _, msg := range *i {
		if msg.Read {
			readCt++
		}
	}
	return readCt
}

func (i *Inbox) Empty() {
	(*i) = Inbox{}
}

func (m Message) DateString() string {
	tFormat := string(configs.GetConfig().TimeFormat)
	return m.DateSent.Format(tFormat)
}
