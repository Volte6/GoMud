package users

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/prompt"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/util"
	//
)

var (
	PermissionGuest string = "guest" // Not logged in
	PermissionUser  string = "user"  // Logged in but no special powers
	PermissionMod   string = "mod"   // Logged in has limited special powers
	PermissionAdmin string = "admin" // Logged in and has special powers

	promptFormat = `<ansi fg="black" bold="true">[</ansi><ansi fg="white">HP:</ansi>` +
		`<ansi fg="hp-%s" bold="%s">` +
		`%d<ansi fg="black" bold="true">/</ansi>%d` +
		`</ansi>` +
		` ` +
		`<ansi fg="white">MP:</ansi>` +
		`<ansi fg="magenta" bold="%s">` +
		`%d<ansi fg="black" bold="true">/</ansi>%d` +
		`</ansi>` +
		`<ansi fg="black" bold="true">]:</ansi>`
)

type UserRecord struct {
	connectionId   uint64
	UserId         int
	Permission     string
	Username       string
	Password       string
	Macros         map[string]string `yaml:"macros,omitempty"` // Up to 10 macros, just string commands.
	Character      *characters.Character
	ItemStorage    Storage `yaml:"itemstorage,omitempty"`
	unsentText     string
	AdminCommands  []string `yaml:"admincommands,omitempty"`
	RoomMemoryBlob string   `yaml:"roommemoryblob,omitempty"`
	configOptions  map[string]any
	connectionTime time.Time
}

func NewUserRecord(userId int, connectionId uint64) *UserRecord {

	return &UserRecord{
		connectionId:   connectionId,
		UserId:         userId,
		Permission:     PermissionGuest,
		Username:       "",
		Password:       "",
		Macros:         make(map[string]string),
		Character:      characters.New(),
		configOptions:  map[string]any{},
		connectionTime: time.Now(),
	}
}

func (u *UserRecord) HasAdminCommand(cmd string) bool {
	if u.Permission != PermissionMod {
		return false
	}

	for _, adminCmd := range u.AdminCommands {
		if adminCmd == cmd {
			return true
		}
	}
	return false
}

func (u *UserRecord) SetConfigOption(key string, value any) {
	u.configOptions[key] = value
}

func (u *UserRecord) GetConfigOption(key string) any {
	if value, ok := u.configOptions[key]; ok {
		return value
	}
	return nil
}

func (u *UserRecord) GetConnectTime() time.Time {
	return u.connectionTime
}

func (u *UserRecord) GetPrompt(fullRedraw bool) string {

	ansiPrompt := ``

	if cmdPrompt := prompt.Get(u.UserId); cmdPrompt != nil {
		if activeQuestion := cmdPrompt.GetNextQuestion(); activeQuestion != nil {
			ansiPrompt = activeQuestion.String()
		}
	}

	if ansiPrompt == `` {

		hpFG := `alive`
		hpBold := `false`
		mpBold := `false`
		if u.Character.Health < 1 {
			hpFG = `dead`
		} else if u.Character.Health == u.Character.HealthMax.Value {
			hpBold = `true`
		}
		if u.Character.Mana == u.Character.ManaMax.Value {
			mpBold = `true`
		}

		ansiPrompt = fmt.Sprintf(promptFormat,
			hpFG, hpBold,
			u.Character.Health, u.Character.HealthMax.Value,
			mpBold,
			u.Character.Mana, u.Character.ManaMax.Value,
		)
	}

	if fullRedraw {
		return term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + ansiPrompt + u.GetUnsentText()
	}

	return ansiPrompt
}

func (u *UserRecord) RoundTick() {

}

// The purpose of AddUnsentText(), ClearUnsentText(), GetUnsentText() is to
// Capture what the user is typing so that when we redraw the
// "prompt" or status bar, we can redraw what they were in the middle
// of typing.
// I don't like the idea of capturing it every time they hit a key though
// There is probably a better way.
func (u *UserRecord) SetUnsentText(t string) {
	u.unsentText = t
}
func (u *UserRecord) ClearUnsentText() {
	u.unsentText = ""
}
func (u *UserRecord) GetUnsentText() string {
	return u.unsentText
}

// Replace a characters information with another.
func (u *UserRecord) ReplaceCharacter(replacement *characters.Character) {
	u.Character = replacement
}

func (u *UserRecord) SetUsername(un string) error {

	if len(un) < minimumUsernameLength || len(un) > maximumUsernameLength {
		return fmt.Errorf("username must be between %d and %d characters long", minimumUsernameLength, maximumUsernameLength)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(un[:1]) {
		return errors.New("username starts with a non alpha character")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(un) {
		return errors.New("username contains non alphanumeric or underscore characters")
	}

	u.Username = un

	// If no character name, just set it to username for now.
	if u.Character.Name == "" {
		u.Character.Name = un
	}

	return nil
}

func (u *UserRecord) SetCharacterName(cn string) error {

	if len(cn) < minimumUsernameLength || len(cn) > maximumUsernameLength {
		return fmt.Errorf("username must be between %d and %d characters long", minimumUsernameLength, maximumUsernameLength)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(cn[:1]) {
		return errors.New("username starts with a non alpha character")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(cn) {
		return errors.New("username contains non alphanumeric or underscore characters")
	}

	u.Character.Name = cn

	return nil
}

func (u *UserRecord) SetPassword(pw string) error {

	if len(pw) < minimumPasswordLength || len(pw) > maximumPasswordLength {
		return fmt.Errorf("password must be between %d and %d characters long", minimumPasswordLength, maximumPasswordLength)
	}

	u.Password = util.Hash(pw)
	return nil
}

func (u *UserRecord) ConnectionId() uint64 {
	return u.connectionId
}
