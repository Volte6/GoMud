package users

import (
	"time"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/items"
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

func (i *Inbox) CountRead() int {
	ct := 0
	for _, msg := range *i {
		if msg.Read {
			ct++
		}
	}
	return ct
}
func (i *Inbox) CountUnread() int {
	ct := 0
	for _, msg := range *i {
		if !msg.Read {
			ct++
		}
	}
	return ct
}

func (i *Inbox) Empty() {
	(*i) = Inbox{}
}

func (m Message) DateString() string {
	tFormat := string(configs.GetConfig().TextFormats.Time)
	return m.DateSent.Format(tFormat)
}
