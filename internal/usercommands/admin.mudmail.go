package usercommands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

func Mudmail(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	/*
		args := util.SplitButRespectQuotes(rest)
		if len(args) < 2 {
			// send some sort of help info?
			infoOutput, _ := templates.Process("admincommands/help/command.mudmail", nil)
			user.SendText(infoOutput)
			return true, nil
		}
	*/

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`mudmail`, rest)
	if isNew {
		user.SendText(fmt.Sprintf(`Starting a new mud mail...%s`, term.CRLFStr))
	}

	msg := users.Message{
		DateSent: time.Now(),
	}

	//
	// From?
	//
	question := cmdPrompt.Ask(`From name?`, []string{})
	if !question.Done {
		return true, nil
	}

	if question.Response == `` {
		user.SendText(`Some name must be provided.`)
		question.RejectResponse()
		return true, nil
	}

	msg.FromName = question.Response

	//
	// Message?
	//
	question = cmdPrompt.Ask(`Message?`, []string{})
	if !question.Done {
		return true, nil
	}

	if question.Response == `` {
		user.ClearPrompt()
		return true, nil
	}

	msg.Message = question.Response

	//
	// Gold?
	//
	question = cmdPrompt.Ask(`Attach how much gold?`, []string{})
	if !question.Done {
		return true, nil
	}

	msg.Gold, _ = strconv.Atoi(question.Response)

	//
	// Attach item?
	//
	question = cmdPrompt.Ask(`Item name (or "none") to attach from your backpack?`, []string{})
	if !question.Done {
		return true, nil
	}

	if question.Response != `none` {
		if itemAttached, found := user.Character.FindInBackpack(question.Response); found {
			msg.Item = &itemAttached
		} else {
			user.SendText(`Could not find item: ` + question.Response)
			question.RejectResponse()
			return true, nil
		}
	}

	//
	// Display preview?
	//
	question = cmdPrompt.Ask(`Send this message to everyone?`, []string{`Yes`, `No`}, `No`)
	if !question.Done {

		tplTxt, _ := templates.Process("mail/message", msg)
		user.SendText(tplTxt)

		return true, nil
	}

	user.ClearPrompt()

	if question.Response[0:1] != `Y` {
		user.SendText(`Okay! Cancelling mass mail.`)
		return true, nil
	}

	users.SearchOfflineUsers(func(u *users.UserRecord) bool {
		u.Inbox.Add(msg)
		users.SaveUser(*u)
		return true
	})

	for _, u := range users.GetAllActiveUsers() {
		u.Inbox.Add(msg)
		users.SaveUser(*u)
		u.Command(`inbox check`)
	}

	user.SendText(``)
	user.SendText(`<ansi fg="alert-5">Message SENT!</ansi>`)
	user.SendText(``)
	return true, nil
}
