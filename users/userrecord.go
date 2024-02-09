package users

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volte6/mud/buffs"
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

	PromptDefault         = `{8}[{255}HP:{hp}{8}/{HP} {255}MP:{13}{mp}{8}/{13}{MP}{8}]{239}{h}{8}:`
	promptDefaultCompiled = CompilePrompt(PromptDefault)
	promptColorRegex      = regexp.MustCompile(`\{(\d*)(?::)?(\d*)?\}`)
	promptFindTagsRegex   = regexp.MustCompile(`\{[a-zA-Z%:\-]+\}`)
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
	suggestText    string
	AdminCommands  []string `yaml:"admincommands,omitempty"`
	RoomMemoryBlob string   `yaml:"roommemoryblob,omitempty"`
	ConfigOptions  map[string]any
	connectionTime time.Time
	lock           sync.RWMutex
	tempDataStore  map[string]any
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
		ConfigOptions:  map[string]any{},
		connectionTime: time.Now(),
		lock:           sync.RWMutex{},
		tempDataStore:  make(map[string]any),
	}
}

func (u *UserRecord) SetTempData(key string, value any) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.tempDataStore == nil {
		u.tempDataStore = make(map[string]any)
	}

	if value == nil {
		delete(u.tempDataStore, key)
		return
	}
	u.tempDataStore[key] = value
}

func (u *UserRecord) GetTempData(key string) any {
	u.lock.RLock()
	defer u.lock.RUnlock()

	if u.tempDataStore == nil {
		u.tempDataStore = make(map[string]any)
	}

	if value, ok := u.tempDataStore[key]; ok {
		return value
	}
	return nil
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
	if u.ConfigOptions == nil {
		u.ConfigOptions = make(map[string]any)
	}
	if value == nil {
		delete(u.ConfigOptions, key)
		return
	}
	u.ConfigOptions[key] = value
}

func (u *UserRecord) GetConfigOption(key string) any {
	if u.ConfigOptions == nil {
		u.ConfigOptions = make(map[string]any)
	}
	if value, ok := u.ConfigOptions[key]; ok {
		return value
	}
	return nil
}

func (u *UserRecord) GetConnectTime() time.Time {
	return u.connectionTime
}

func (u *UserRecord) GetPrompt(fullRedraw bool) string {

	u.lock.RLock()
	defer u.lock.RUnlock()

	ansiPrompt := ``

	if cmdPrompt := prompt.Get(u.UserId); cmdPrompt != nil {
		if activeQuestion := cmdPrompt.GetNextQuestion(); activeQuestion != nil {
			ansiPrompt = activeQuestion.String()
		}
	}

	if ansiPrompt == `` {

		/*

			mpBold := `false`
			if u.Character.Mana == u.Character.ManaMax.Value {
				mpBold = `true`
			}
				ansiPrompt = fmt.Sprintf(promptFormat,
					util.HealthClass(u.Character.Health, u.Character.HealthMax.Value),
					u.Character.Health, u.Character.HealthMax.Value,
					mpBold,
					u.Character.Mana, u.Character.ManaMax.Value,
				)
		*/

		var customPrompt any = nil
		var inCombat bool = u.Character.Aggro != nil

		if inCombat {
			customPrompt = u.GetConfigOption(`fprompt-compiled`)
		}

		// No other custom prompts? try the default setting
		if customPrompt == nil {
			customPrompt = u.GetConfigOption(`prompt-compiled`)
		}

		var ok bool
		if customPrompt == nil {
			ansiPrompt = promptDefaultCompiled
		} else if ansiPrompt, ok = customPrompt.(string); !ok {
			ansiPrompt = promptDefaultCompiled
		}

		//
		// TODO: Need to optimize this section to only calculate/replace when the value is actually used.
		//
		var currentXP, tnlXP int = -1, -1
		var hpPct, mpPct int = -1, -1
		var hpClass, mpClass string

		matches := promptFindTagsRegex.FindAllString(ansiPrompt, -1)
		for _, match := range matches {

			switch match {

			case "{hp}":
				if len(hpClass) == 0 {
					hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{hp}", fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, hpClass, u.Character.Health))

			case "{hp:-}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{hp:-}", strconv.Itoa(u.Character.Health))

			case "{HP}":
				if len(hpClass) == 0 {
					hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{HP}", fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, hpClass, u.Character.HealthMax.Value))

			case "{HP:-}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{HP:-}", strconv.Itoa(u.Character.HealthMax.Value))

			case "{hp%}":
				if hpPct == -1 {
					hpPct = int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
				}
				if len(hpClass) == 0 {
					hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{hp%}", fmt.Sprintf(`<ansi fg="%s">%d%%</ansi>`, hpClass, hpPct))

			case "{hp%:-}":
				if hpPct == -1 {
					hpPct = int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{hp%:-}", strconv.Itoa(hpPct)+`%`)

			case "{mp}":
				if len(mpClass) == 0 {
					mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{mp}", fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, mpClass, u.Character.Mana))

			case "{mp:-}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{mp:-}", strconv.Itoa(u.Character.Mana))

			case "{MP}":
				if len(mpClass) == 0 {
					mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{MP}", fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, mpClass, u.Character.ManaMax.Value))

			case "{MP:-}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{MP:-}", strconv.Itoa(u.Character.ManaMax.Value))

			case "{mp%}":
				if mpPct == -1 {
					mpPct = int(math.Floor(float64(u.Character.Mana) / float64(u.Character.ManaMax.Value) * 100))
				}
				if len(mpClass) == 0 {
					mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{mp%}", fmt.Sprintf(`<ansi fg="%s">%d%%</ansi>`, mpClass, mpPct))

			case "{mp%:-}":
				if mpPct == -1 {
					mpPct = int(math.Floor(float64(u.Character.Mana) / float64(u.Character.ManaMax.Value) * 100))
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{mp%:-}", strconv.Itoa(mpPct)+`%`)

			case "{xp}":
				if currentXP == -1 && tnlXP == -1 {
					currentXP, tnlXP = u.Character.XPTNLActual()
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{xp}", strconv.Itoa(currentXP))

			case "{XP}":
				if currentXP == -1 && tnlXP == -1 {
					currentXP, tnlXP = u.Character.XPTNLActual()
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{XP}", strconv.Itoa(tnlXP))

			case "{xp%}":
				if currentXP == -1 && tnlXP == -1 {
					currentXP, tnlXP = u.Character.XPTNLActual()
				}
				tnlPercent := int(math.Floor(float64(currentXP) / float64(tnlXP) * 100))
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{xp%}", strconv.Itoa(tnlPercent)+`%`)

			case "{h}":
				hiddenFlag := ``
				if u.Character.HasBuffFlag(buffs.Hidden) {
					hiddenFlag = `H`
				}
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{h}", hiddenFlag)

			case "{a}":
				alignClass := u.Character.AlignmentName()
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{a}", fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignClass, alignClass[:1]))

			case "{A}":
				alignClass := u.Character.AlignmentName()
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{A}", fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignClass, alignClass))

			case "{g}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{g}", strconv.Itoa(u.Character.Gold))

			case "{tp}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{tp}", strconv.Itoa(u.Character.TrainingPoints))

			case "{sp}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{sp}", strconv.Itoa(u.Character.StatPoints))

			case "{i}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{i}", strconv.Itoa(len(u.Character.Items)))

			case "{I}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{I}", strconv.Itoa(u.Character.GetBackpackCapacity()))

			case "{lvl}":
				ansiPrompt = strings.ReplaceAll(ansiPrompt, "{lvl}", strconv.Itoa(u.Character.Level))

			case "{w}":
				if inCombat {
					ansiPrompt = strings.ReplaceAll(ansiPrompt, "{w}", strconv.Itoa(u.Character.Aggro.RoundsWaiting))
				}
			}
		}

	}

	if fullRedraw {
		unsent, suggested := u.GetUnsentText()
		if len(suggested) > 0 {
			suggested = `<ansi fg="suggested-text">` + suggested + `</ansi>`
		}
		return term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + ansiPrompt + unsent + suggested
	}

	return ansiPrompt
}

func CompilePrompt(input string) string {

	if promptColorRegex.MatchString(input) {
		input = `<ansi bg="" fg="">` + promptColorRegex.ReplaceAllString(input, `</ansi><ansi fg="$1" bg="$2">`) + `</ansi>`
	}

	input = strings.ReplaceAll(input, ` bg=""`, ``)
	input = strings.ReplaceAll(input, ` fg=""`, ``)
	input = strings.ReplaceAll(input, `<ansi></ansi>`, ``)

	return input
}

func (u *UserRecord) RoundTick() {

}

// The purpose of SetUnsentText(), GetUnsentText() is to
// Capture what the user is typing so that when we redraw the
// "prompt" or status bar, we can redraw what they were in the middle
// of typing.
// I don't like the idea of capturing it every time they hit a key though
// There is probably a better way.
func (u *UserRecord) SetUnsentText(t string, suggest string) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.unsentText = t
	u.suggestText = suggest
}

func (u *UserRecord) GetUnsentText() (unsent string, suggestion string) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.unsentText, u.suggestText
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
