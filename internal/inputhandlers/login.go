package inputhandlers

import (
	"log/slog"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

type LoginState struct {
	SentWelcome      bool
	PasswordAttempts int
	UserObject       *users.UserRecord
}

func LoginInputHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	usernamePrompt, _ := templates.Process("login/username.prompt", nil)
	passwordPrompt, _ := templates.Process("login/password.prompt", nil)
	passwordMask, _ := templates.Process("login/password.mask", nil)

	if !connections.IsWebsocket(clientInput.ConnectionId) {
		usernamePrompt = templates.AnsiParse(usernamePrompt)
		passwordPrompt = templates.AnsiParse(passwordPrompt)
		passwordMask = templates.AnsiParse(passwordMask)
	}

	var state *LoginState

	if val, ok := sharedState["LoginInputHandler"]; !ok {
		state = &LoginState{
			SentWelcome:      false,
			PasswordAttempts: 0,
			UserObject:       users.NewUserRecord(0, clientInput.ConnectionId),
		}
		sharedState["LoginInputHandler"] = state
	} else {
		state = val.(*LoginState)
	}

	if !state.SentWelcome {
		state.SentWelcome = true
		splashTxt, _ := templates.Process("login/connect-splash", nil)

		if !connections.IsWebsocket(clientInput.ConnectionId) {
			splashTxt = templates.AnsiParse(splashTxt)
		}

		connections.SendTo([]byte(splashTxt), clientInput.ConnectionId)
		connections.SendTo([]byte(usernamePrompt), clientInput.ConnectionId)
	}

	if len(state.UserObject.Username) > 0 && len(state.UserObject.Password) < 1 {
		// passwords we only sent back a * for each character
		for i := 0; i < len(clientInput.DataIn); i++ {
			connections.SendTo([]byte(passwordMask), clientInput.ConnectionId)
		}
	} else {
		// Everything else gets echoed back normally.
		connections.SendTo(clientInput.DataIn, clientInput.ConnectionId)
	}
	// We only care about processing input after they hit enter.
	if !clientInput.EnterPressed {
		return false
	}

	//
	// If we've reached this point they hit enter.
	//

	// Special case to check up front if they just hit enter with no input.
	// If waiting on the y/n answer, default to "n"
	// maybe refactor some of this later.
	if len(state.UserObject.Username) > 0 && len(state.UserObject.Password) > 0 && state.UserObject.UserId == 0 {
		if len(clientInput.Buffer) < 1 {
			clientInput.DataIn = []byte("no")
			connections.SendTo(clientInput.DataIn, clientInput.ConnectionId)
		}
	}

	connections.SendTo(term.CRLF, clientInput.ConnectionId)

	submittedText := make([]byte, len(clientInput.Buffer))
	copy(submittedText, clientInput.Buffer)
	clientInput.Buffer = []byte{}

	// If they haven't submitted a username yet, we need to process that.
	if len(state.UserObject.Username) < 1 {
		if err := state.UserObject.SetUsername(string(submittedText)); err != nil {
			connections.SendTo([]byte(err.Error()), clientInput.ConnectionId)    // error message
			connections.SendTo(term.CRLF, clientInput.ConnectionId)              // Newline
			connections.SendTo([]byte(usernamePrompt), clientInput.ConnectionId) // prompt
			return false
		}

		// Setting username was a success, send the password prompt
		connections.SendTo([]byte(passwordPrompt), clientInput.ConnectionId)

		events.AddToQueue(events.WebClientCommand{
			ConnectionId: clientInput.ConnectionId,
			Text:         `TEXTMASK:true`,
		})

		return false
	}

	if len(state.UserObject.Password) < 1 {

		if err := state.UserObject.SetPassword(string(submittedText)); err != nil {
			connections.SendTo([]byte(err.Error()), clientInput.ConnectionId)    // error message
			connections.SendTo(term.CRLF, clientInput.ConnectionId)              // Newline
			connections.SendTo([]byte(passwordPrompt), clientInput.ConnectionId) // prompt
			return false
		}

		if users.Exists(state.UserObject.Username) {

			tmpUser, err := users.LoadUser(state.UserObject.Username)
			if err != nil {
				panic(err)
			} else if !tmpUser.PasswordMatches(state.UserObject.Password) {
				connections.SendTo([]byte("Oops, bye!"), clientInput.ConnectionId)
				connections.SendTo(term.CRLF, clientInput.ConnectionId) // Newline
				connections.Remove(clientInput.ConnectionId)
			} else {

				events.AddToQueue(events.WebClientCommand{
					ConnectionId: clientInput.ConnectionId,
					Text:         `TEXTMASK:false`,
				})

				tmpUser, msg, err := users.LoginUser(tmpUser, clientInput.ConnectionId)

				// Password matched, assign the loaded data
				if tmpUser != nil {
					state.UserObject = tmpUser
				}

				if len(msg) > 0 {
					connections.SendTo([]byte(msg), clientInput.ConnectionId)
					connections.SendTo(term.CRLF, clientInput.ConnectionId) // Newline
				}

				if err != nil {
					connections.Remove(clientInput.ConnectionId)
					return false
				}

				return true
			}

		} else {

			events.AddToQueue(events.WebClientCommand{
				ConnectionId: clientInput.ConnectionId,
				Text:         `TEXTMASK:false`,
			})

			newUserPromptPrompt, _ := templates.Process("generic/prompt.yn", map[string]any{
				"prompt":  "Would you like to create a new user?",
				"options": []string{"y", "n"},
				"default": "n",
			})

			if !connections.IsWebsocket(clientInput.ConnectionId) {
				newUserPromptPrompt = templates.AnsiParse(newUserPromptPrompt)
			}
			connections.SendTo([]byte(newUserPromptPrompt), clientInput.ConnectionId)
		}

		return false

	}

	// If no user id, must be a new user.
	if len(submittedText) < 1 {
		submittedText = []byte("n")

	}
	if submittedText[0] != 'y' && submittedText[0] != 'Y' {
		connections.SendTo([]byte("Oops, bye!"), clientInput.ConnectionId)
		connections.SendTo(term.CRLF, clientInput.ConnectionId) // Newline
		connections.Remove(clientInput.ConnectionId)
		return false
	}

	if err := users.CreateUser(state.UserObject); err != nil {
		slog.Error("Could not create user", "error", err.Error())

		connections.SendTo([]byte("Could not create user: "+err.Error()+"\n"), clientInput.ConnectionId)

		connections.SendTo([]byte("Oops, bye!"), clientInput.ConnectionId)
		connections.SendTo(term.CRLF, clientInput.ConnectionId) // Newline
		connections.Remove(clientInput.ConnectionId)
		return false
	}

	// Once complete, return true to let main.go know we're done with this handler.
	return true

}
