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
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/progressbar"
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

	PromptDefault         = `{8}[{t} {T} {255}HP:{hp}{8}/{HP} {255}MP:{13}{mp}{8}/{13}{MP}{8}]{239}{h}{8}:`
	promptDefaultCompiled = util.ConvertColorShortTags(PromptDefault)
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
	activePrompt   *prompt.Prompt
	progress       *progressbar.ProgressBar
	isZombie       bool // are they a zombie currently?
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

func (u *UserRecord) PasswordMatches(input string) bool {

	if input == u.Password {
		return true
	}

	// In case we reset the password to a plaintext string
	if input == util.Hash(u.Password) {
		return true
	}

	return false
}

func (u *UserRecord) SetProgressBar(pb *progressbar.ProgressBar) {
	u.progress = pb
}

func (u *UserRecord) GetProgressBar() *progressbar.ProgressBar {
	return u.progress
}

func (u *UserRecord) RemoveProgressBar() {
	u.progress.OnComplete()
	u.progress = nil
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

func (u *UserRecord) GetCommandPrompt(fullRedraw bool) string {

	u.lock.RLock()
	defer u.lock.RUnlock()

	promptOut := strings.Builder{}

	promptPrefix := ``
	promptSuffix := ``

	if u.activePrompt != nil {

		if activeQuestion := u.activePrompt.GetNextQuestion(); activeQuestion != nil {
			promptOut.WriteString(activeQuestion.String())
		}
	}

	if u.progress != nil {

		rStyle := u.progress.RenderStyle()

		if rStyle == progressbar.PromptReplace {
			promptOut.WriteString(u.progress.String())
		} else if rStyle == progressbar.PromptPrefix {
			promptPrefix = u.progress.String()
		} else if rStyle == progressbar.PromptSuffix {
			promptSuffix = u.progress.String()
		}

	}

	if promptOut.Len() == 0 {

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
		ansiPrompt := ``
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
		var acPt int = u.Character.ActionPoints
		var hpClass, mpClass string

		promptLen := len(ansiPrompt)
		tagStartPos := -1

		for i := 0; i < promptLen; i++ {
			if ansiPrompt[i] == '{' {
				tagStartPos = i
				continue
			}
			if ansiPrompt[i] == '}' {

				switch ansiPrompt[tagStartPos : i+1] {

				case "{hp}":
					if len(hpClass) == 0 {
						hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
					}
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, hpClass, u.Character.Health))

				case "{hp:-}":
					promptOut.WriteString(strconv.Itoa(u.Character.Health))
				case "{HP}":
					if len(hpClass) == 0 {
						hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
					}
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, hpClass, u.Character.HealthMax.Value))
				case "{HP:-}":
					promptOut.WriteString(strconv.Itoa(u.Character.HealthMax.Value))
				case "{hp%}":
					if hpPct == -1 {
						hpPct = int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
					}
					if len(hpClass) == 0 {
						hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
					}
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d%%</ansi>`, hpClass, hpPct))

				case "{hp%:-}":
					if hpPct == -1 {
						hpPct = int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
					}
					promptOut.WriteString(strconv.Itoa(hpPct))
					promptOut.WriteString(`%`)

				case "{mp}":
					if len(mpClass) == 0 {
						mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
					}
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, mpClass, u.Character.Mana))

				case "{mp:-}":
					promptOut.WriteString(strconv.Itoa(u.Character.Mana))

				case "{MP}":
					if len(mpClass) == 0 {
						mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
					}
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, mpClass, u.Character.ManaMax.Value))

				case "{MP:-}":
					promptOut.WriteString(strconv.Itoa(u.Character.ManaMax.Value))

				case "{mp%}":
					if mpPct == -1 {
						mpPct = int(math.Floor(float64(u.Character.Mana) / float64(u.Character.ManaMax.Value) * 100))
					}
					if len(mpClass) == 0 {
						mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
					}
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d%%</ansi>`, mpClass, mpPct))

				case "{mp%:-}":
					if mpPct == -1 {
						mpPct = int(math.Floor(float64(u.Character.Mana) / float64(u.Character.ManaMax.Value) * 100))
					}
					promptOut.WriteString(strconv.Itoa(mpPct))
					promptOut.WriteString(`%`)

				case "{ap}":
					promptOut.WriteString(strconv.Itoa(acPt))

				case "{xp}":
					if currentXP == -1 && tnlXP == -1 {
						currentXP, tnlXP = u.Character.XPTNLActual()
					}
					promptOut.WriteString(strconv.Itoa(currentXP))

				case "{XP}":
					if currentXP == -1 && tnlXP == -1 {
						currentXP, tnlXP = u.Character.XPTNLActual()
					}
					promptOut.WriteString(strconv.Itoa(tnlXP))

				case "{xp%}":
					if currentXP == -1 && tnlXP == -1 {
						currentXP, tnlXP = u.Character.XPTNLActual()
					}
					tnlPercent := int(math.Floor(float64(currentXP) / float64(tnlXP) * 100))
					promptOut.WriteString(strconv.Itoa(tnlPercent))
					promptOut.WriteString(`%`)

				case "{h}":
					hiddenFlag := ``
					if u.Character.HasBuffFlag(buffs.Hidden) {
						hiddenFlag = `H`
					}
					promptOut.WriteString(hiddenFlag)

				case "{a}":
					alignClass := u.Character.AlignmentName()
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignClass, alignClass[:1]))

				case "{A}":
					alignClass := u.Character.AlignmentName()
					promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignClass, alignClass))

				case "{g}":
					promptOut.WriteString(strconv.Itoa(u.Character.Gold))

				case "{tp}":
					promptOut.WriteString(strconv.Itoa(u.Character.TrainingPoints))

				case "{sp}":
					promptOut.WriteString(strconv.Itoa(u.Character.StatPoints))

				case "{i}":
					promptOut.WriteString(strconv.Itoa(len(u.Character.Items)))

				case "{I}":
					promptOut.WriteString(strconv.Itoa(u.Character.GetBackpackCapacity()))

				case "{lvl}":
					promptOut.WriteString(strconv.Itoa(u.Character.Level))

				case "{w}":
					if inCombat {
						promptOut.WriteString(strconv.Itoa(u.Character.Aggro.RoundsWaiting))
					}

				case "{t}":
					gd := gametime.GetDate()
					promptOut.WriteString(gd.String(true))

				case "{T}":
					gd := gametime.GetDate()
					promptOut.WriteString(gd.String())

				}
				tagStartPos = -1
				continue
			}

			if tagStartPos == -1 {
				promptOut.WriteByte(ansiPrompt[i])
			}
		}

	}

	if fullRedraw {
		unsent, suggested := u.GetUnsentText()
		if len(suggested) > 0 {
			suggested = `<ansi fg="suggested-text">` + suggested + `</ansi>`
		}
		return term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + promptPrefix + promptOut.String() + promptSuffix + unsent + suggested
	}

	return promptPrefix + promptOut.String() + promptSuffix
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

// Prompt related functionality
func (u *UserRecord) StartPrompt(command string, rest string) (*prompt.Prompt, bool) {

	u.lock.RLock()
	defer u.lock.RUnlock()

	if u.activePrompt != nil {
		// If it's the same prompt, return the existing one
		if u.activePrompt.Command == command && u.activePrompt.Rest == rest {
			return u.activePrompt, false
		}
	}

	// If no prompt found or it seems like a new prompt, create a new one and replace the old
	u.activePrompt = prompt.New(command, rest)

	return u.activePrompt, true
}

func (u *UserRecord) GetPrompt() *prompt.Prompt {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.activePrompt
}

func (u *UserRecord) ClearPrompt() {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.activePrompt = nil
}
