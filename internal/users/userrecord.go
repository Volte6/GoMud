package users

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/prompt"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/util"
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
	UserId         int                   `yaml:"userid"`
	Permission     string                `yaml:"permission"`
	Username       string                `yaml:"username"`
	Password       string                `yaml:"password"`
	Joined         time.Time             `yaml:"joined"`
	Macros         map[string]string     `yaml:"macros,omitempty"` // Up to 10 macros, just string commands.
	Character      *characters.Character `yaml:"character,omitempty"`
	ItemStorage    Storage               `yaml:"itemstorage,omitempty"`
	AdminCommands  []string              `yaml:"admincommands,omitempty"`
	RoomMemoryBlob string                `yaml:"roommemoryblob,omitempty"`
	ConfigOptions  map[string]any        `yaml:"configoptions,omitempty"`
	Inbox          Inbox                 `yaml:"inbox,omitempty"`
	Muted          bool                  `yaml:"muted,omitempty"`    // Cannot SEND custom communications to anyone but admin/mods
	Deafened       bool                  `yaml:"deafened,omitempty"` // Cannot HEAR custom communications from anyone but admin/mods
	connectionId   uint64
	unsentText     string
	suggestText    string
	connectionTime time.Time
	lastInputRound uint64
	tempDataStore  map[string]any
	activePrompt   *prompt.Prompt
	isZombie       bool // are they a zombie currently?
}

func NewUserRecord(userId int, connectionId uint64) *UserRecord {

	c := configs.GetConfig()

	u := &UserRecord{
		connectionId:   connectionId,
		UserId:         userId,
		Permission:     PermissionGuest,
		Username:       "",
		Password:       "",
		Macros:         make(map[string]string),
		Character:      characters.New(),
		ConfigOptions:  map[string]any{},
		Joined:         time.Now(),
		connectionTime: time.Now(),
		tempDataStore:  make(map[string]any),
	}

	if c.PermaDeath {
		u.Character.ExtraLives = int(c.LivesStart)
	}

	return u
}

func (u *UserRecord) ClientSettings() connections.ClientSettings {
	return connections.GetClientSettings(u.connectionId)
}

func (u *UserRecord) PasswordMatches(input string) bool {

	if input == u.Password {
		return true
	}

	if u.Password == util.Hash(input) {
		return true
	}

	// In case we reset the password to a plaintext string
	if input == util.Hash(u.Password) {
		return true
	}

	return false
}

func (u *UserRecord) ShorthandId() string {
	return fmt.Sprintf(`@%d`, u.UserId)
}

func (u *UserRecord) SetLastInputRound(rdNum uint64) {
	u.lastInputRound = rdNum
}

func (u *UserRecord) GetLastInputRound() uint64 {
	return u.lastInputRound
}

func (u *UserRecord) HasShop() bool {
	return len(u.Character.Shop) > 0
}

func (u *UserRecord) Command(inputTxt string, waitTurns ...int) {

	wt := 0
	if len(waitTurns) > 0 {
		wt = waitTurns[0]
	}

	events.AddToQueue(events.Input{
		UserId:    u.UserId,
		InputText: inputTxt,
		WaitTurns: wt,
	})

}

func (u *UserRecord) AddBuff(buffId int) {

	events.AddToQueue(events.Buff{
		UserId: u.UserId,
		BuffId: buffId,
	})

}

func (u *UserRecord) SendText(txt string) {

	events.AddToQueue(events.Message{
		UserId: u.UserId,
		Text:   txt + "\n",
	})

}

func (u *UserRecord) SendWebClientCommand(txt string) {

	events.AddToQueue(events.WebClientCommand{
		ConnectionId: u.connectionId,
		Text:         txt,
	})

}

func (u *UserRecord) SetTempData(key string, value any) {

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

func (u *UserRecord) GetCommandPrompt(fullRedraw bool, forcePromptType ...string) string {

	promptOut := ``

	if len(forcePromptType) == 0 || forcePromptType[0] != `mprompt` {
		if u.activePrompt != nil {

			if activeQuestion := u.activePrompt.GetNextQuestion(); activeQuestion != nil {
				promptOut = activeQuestion.String()
			}
		}
	}

	if len(promptOut) == 0 {

		var customPrompt any = nil
		var inCombat bool = u.Character.Aggro != nil

		if len(forcePromptType) > 0 {
			customPrompt = u.GetConfigOption(forcePromptType[0] + `-compiled`)
		} else {

			if inCombat {
				customPrompt = u.GetConfigOption(`fprompt-compiled`)
			}

			// No other custom prompts? try the default setting
			if customPrompt == nil {
				customPrompt = u.GetConfigOption(`prompt-compiled`)
			}
		}

		var ok bool
		ansiPrompt := ``
		if customPrompt == nil {
			ansiPrompt = promptDefaultCompiled
		} else if ansiPrompt, ok = customPrompt.(string); !ok {
			ansiPrompt = promptDefaultCompiled
		}

		promptOut = u.ProcessPromptString(ansiPrompt)

	}

	if fullRedraw {
		unsent, suggested := u.GetUnsentText()
		if len(suggested) > 0 {
			suggested = `<ansi fg="suggested-text">` + suggested + `</ansi>`
		}
		return term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + promptOut + unsent + suggested
	}

	return promptOut
}

func (u *UserRecord) ProcessPromptString(promptStr string) string {

	promptOut := strings.Builder{}

	var currentXP, tnlXP int = -1, -1
	var hpPct, mpPct int = -1, -1
	var hpClass, mpClass string

	promptLen := len(promptStr)
	tagStartPos := -1

	for i := 0; i < promptLen; i++ {
		if promptStr[i] == '{' {
			tagStartPos = i
			continue
		}
		if promptStr[i] == '}' {

			switch promptStr[tagStartPos : i+1] {

			case `{\n}`:
				promptOut.WriteString("\n")

			case `{hp}`:
				if len(hpClass) == 0 {
					hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
				}
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, hpClass, u.Character.Health))

			case `{hp:-}`:
				promptOut.WriteString(strconv.Itoa(u.Character.Health))
			case `{HP}`:
				if len(hpClass) == 0 {
					hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
				}
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, hpClass, u.Character.HealthMax.Value))
			case `{HP:-}`:
				promptOut.WriteString(strconv.Itoa(u.Character.HealthMax.Value))
			case `{hp%}`:
				if hpPct == -1 {
					hpPct = int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
				}
				if len(hpClass) == 0 {
					hpClass = fmt.Sprintf(`health-%d`, util.QuantizeTens(u.Character.Health, u.Character.HealthMax.Value))
				}
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d%%</ansi>`, hpClass, hpPct))

			case `{hp%:-}`:
				if hpPct == -1 {
					hpPct = int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
				}
				promptOut.WriteString(strconv.Itoa(hpPct))
				promptOut.WriteString(`%`)

			case `{mp}`:
				if len(mpClass) == 0 {
					mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
				}
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, mpClass, u.Character.Mana))

			case `{mp:-}`:
				promptOut.WriteString(strconv.Itoa(u.Character.Mana))

			case `{MP}`:
				if len(mpClass) == 0 {
					mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
				}
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d</ansi>`, mpClass, u.Character.ManaMax.Value))

			case `{MP:-}`:
				promptOut.WriteString(strconv.Itoa(u.Character.ManaMax.Value))

			case `{mp%}`:
				if mpPct == -1 {
					mpPct = int(math.Floor(float64(u.Character.Mana) / float64(u.Character.ManaMax.Value) * 100))
				}
				if len(mpClass) == 0 {
					mpClass = fmt.Sprintf(`mana-%d`, util.QuantizeTens(u.Character.Mana, u.Character.ManaMax.Value))
				}
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%d%%</ansi>`, mpClass, mpPct))

			case `{mp%:-}`:
				if mpPct == -1 {
					mpPct = int(math.Floor(float64(u.Character.Mana) / float64(u.Character.ManaMax.Value) * 100))
				}
				promptOut.WriteString(strconv.Itoa(mpPct))
				promptOut.WriteString(`%`)

			case `{ap}`:
				promptOut.WriteString(strconv.Itoa(u.Character.ActionPoints))

			case `{xp}`:
				if currentXP == -1 && tnlXP == -1 {
					currentXP, tnlXP = u.Character.XPTNLActual()
				}
				promptOut.WriteString(strconv.Itoa(currentXP))

			case `{XP}`:
				if currentXP == -1 && tnlXP == -1 {
					currentXP, tnlXP = u.Character.XPTNLActual()
				}
				promptOut.WriteString(strconv.Itoa(tnlXP))

			case `{xp%}`:
				if currentXP == -1 && tnlXP == -1 {
					currentXP, tnlXP = u.Character.XPTNLActual()
				}
				tnlPercent := int(math.Floor(float64(currentXP) / float64(tnlXP) * 100))
				promptOut.WriteString(strconv.Itoa(tnlPercent))
				promptOut.WriteString(`%`)

			case `{h}`:
				hiddenFlag := ``
				if u.Character.HasBuffFlag(buffs.Hidden) {
					hiddenFlag = `H`
				}
				promptOut.WriteString(hiddenFlag)

			case `{a}`:
				alignClass := u.Character.AlignmentName()
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignClass, alignClass[:1]))

			case `{A}`:
				alignClass := u.Character.AlignmentName()
				promptOut.WriteString(fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignClass, alignClass))

			case `{g}`:
				promptOut.WriteString(strconv.Itoa(u.Character.Gold))

			case `{tp}`:
				promptOut.WriteString(strconv.Itoa(u.Character.TrainingPoints))

			case `{sp}`:
				promptOut.WriteString(strconv.Itoa(u.Character.StatPoints))

			case `{i}`:
				promptOut.WriteString(strconv.Itoa(len(u.Character.Items)))

			case `{I}`:
				promptOut.WriteString(strconv.Itoa(u.Character.CarryCapacity()))

			case `{lvl}`:
				promptOut.WriteString(strconv.Itoa(u.Character.Level))

			case `{w}`:
				if u.Character.Aggro != nil {
					promptOut.WriteString(strconv.Itoa(u.Character.Aggro.RoundsWaiting))
				} else {
					promptOut.WriteString(`0`)
				}

			case `{t}`:
				gd := gametime.GetDate()
				promptOut.WriteString(gd.String(true))

			case `{T}`:
				gd := gametime.GetDate()
				promptOut.WriteString(gd.String())

			}
			tagStartPos = -1
			continue
		}

		if tagStartPos == -1 {
			promptOut.WriteByte(promptStr[i])
		}
	}

	return promptOut.String()
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

	u.unsentText = t
	u.suggestText = suggest
}

func (u *UserRecord) GetUnsentText() (unsent string, suggestion string) {

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

	return u.activePrompt
}

func (u *UserRecord) ClearPrompt() {

	u.activePrompt = nil
}

func (u *UserRecord) GetOnlineInfo() OnlineInfo {

	c := configs.GetConfig()
	afkRounds := uint64(c.SecondsToRounds(int(c.AfkSeconds)))
	roundNow := util.GetRoundCount()

	connTime := u.GetConnectTime()

	oTime := time.Since(connTime)

	h := int(math.Floor(oTime.Hours()))
	m := int(math.Floor(oTime.Minutes())) - (h * 60)
	s := int(math.Floor(oTime.Seconds())) - (h * 60 * 60) - (m * 60)

	timeStr := ``
	if h > 0 {
		timeStr = fmt.Sprintf(`%dh%dm`, h, m)
	} else if m > 0 {
		timeStr = fmt.Sprintf(`%dm`, m)
	} else {
		timeStr = fmt.Sprintf(`%ds`, s)
	}

	isAfk := false
	if afkRounds > 0 && roundNow-u.GetLastInputRound() >= afkRounds {
		isAfk = true
	}

	return OnlineInfo{
		u.Username,
		u.Character.Name,
		u.Character.Level,
		u.Character.AlignmentName(),
		skills.GetProfession(u.Character.GetAllSkillRanks()),
		int64(oTime.Seconds()),
		timeStr,
		isAfk,
		u.Permission,
	}
}
