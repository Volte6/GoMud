package users

import "time"

type Inbox []Message

type Message struct {
	FromUserId int
	FromName   string
	Message    string
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
