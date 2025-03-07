package users

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/volte6/gomud/internal/audio"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/prompt"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/util"
	//
)

var (
	PermissionGuest string = "guest" // Not logged in
	PermissionUser  string = "user"  // Logged in but no special powers
	PermissionMod   string = "mod"   // Logged in has limited special powers
	PermissionAdmin string = "admin" // Logged in and has special powers
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
	ConfigOptions  map[string]any        `yaml:"configoptions,omitempty"`
	Inbox          Inbox                 `yaml:"inbox,omitempty"`
	Muted          bool                  `yaml:"muted,omitempty"`    // Cannot SEND custom communications to anyone but admin/mods
	Deafened       bool                  `yaml:"deafened,omitempty"` // Cannot HEAR custom communications from anyone but admin/mods
	EventLog       UserLog               `yaml:"-"`                  // Do not retain in user file (for now)
	LastMusic      string                `yaml:"-"`                  // Keeps track of the last music that was played
	connectionId   uint64
	unsentText     string
	suggestText    string
	connectionTime time.Time
	lastInputRound uint64
	tempDataStore  map[string]any
	activePrompt   *prompt.Prompt
	isZombie       bool // are they a zombie currently?
	inputBlocked   bool // Whether input is currently intentionally turned off (for a certain category of commands)
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
		EventLog:       UserLog{},
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

// Grants experience to the user and notifies them
// Additionally accepts `source` as a short identifier of the XP source
// Example source: "combat", "quest progress", "trash cleanup", "exploration"
func (u *UserRecord) GrantXP(amt int, source string) {

	grantXP, xpScale := u.Character.GrantXP(amt)

	if xpScale != 100 {
		u.SendText(fmt.Sprintf(`You gained <ansi fg="yellow-bold">%d experience points</ansi> <ansi fg="yellow">(%d%% scale)</ansi>! <ansi fg="7">(%s)</ansi>`, grantXP, xpScale, source))

		u.EventLog.Add(`xp`, fmt.Sprintf(`Gained <ansi fg="yellow-bold">%d experience points</ansi> <ansi fg="yellow">(%d%% scale)</ansi>! <ansi fg="7">(%s)</ansi>`, grantXP, xpScale, source))

	} else {

		u.SendText(fmt.Sprintf(`You gained <ansi fg="yellow-bold">%d experience points</ansi>! <ansi fg="7">(%s)</ansi>`, grantXP, source))

		u.EventLog.Add(`xp`, fmt.Sprintf(`Gained <ansi fg="yellow-bold">%d experience points</ansi>! <ansi fg="7">(%s)</ansi>`, grantXP, source))
	}

	newLevel, statsDelta := u.Character.LevelUp()
	for newLevel {

		c := configs.GetConfig()

		livesBefore := u.Character.ExtraLives

		if c.PermaDeath && c.LivesOnLevelUp > 0 {
			u.Character.ExtraLives += int(c.LivesOnLevelUp)
			if u.Character.ExtraLives > int(c.LivesMax) {
				u.Character.ExtraLives = int(c.LivesMax)
			}
		}

		u.EventLog.Add(`xp`, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now <ansi fg="magenta-bold">level %d</ansi>!`, u.Character.Name, u.Character.Level))

		SaveUser(*u)

		events.AddToQueue(events.LevelUp{
			UserId:         u.UserId,
			RoomId:         u.Character.RoomId,
			Username:       u.Username,
			CharacterName:  u.Character.Name,
			NewLevel:       u.Character.Level,
			StatsDelta:     statsDelta,
			TrainingPoints: 1,
			StatPoints:     1,
			LivesGained:    u.Character.ExtraLives - livesBefore,
		})

		newLevel, statsDelta = u.Character.LevelUp()
	}
}

func (u *UserRecord) PlayMusic(musicFileOrId string) {

	v := 100
	if soundConfig := audio.GetFile(musicFileOrId); soundConfig.FilePath != `` {
		musicFileOrId = soundConfig.FilePath
		if soundConfig.Volume > 0 && soundConfig.Volume <= 100 {
			v = soundConfig.Volume
		}
	}

	events.AddToQueue(events.MSP{
		UserId:    u.UserId,
		SoundType: `MUSIC`,
		SoundFile: musicFileOrId,
		Volume:    v,
	})

}

func (u *UserRecord) PlaySound(soundId string, category string) {

	v := 100
	if soundConfig := audio.GetFile(soundId); soundConfig.FilePath != `` {
		soundId = soundConfig.FilePath
		if soundConfig.Volume > 0 && soundConfig.Volume <= 100 {
			v = soundConfig.Volume
		}
	}

	events.AddToQueue(events.MSP{
		UserId:    u.UserId,
		SoundType: `SOUND`,
		SoundFile: soundId,
		Volume:    v,
		Category:  category,
	})

}

func (u *UserRecord) Command(inputTxt string, waitSeconds ...float64) {

	readyTurn := util.GetTurnCount()
	if len(waitSeconds) > 0 {
		readyTurn += uint64(float64(configs.GetConfig().SecondsToTurns(1)) * waitSeconds[0])
	}

	events.AddToQueue(events.Input{
		UserId:    u.UserId,
		InputText: inputTxt,
		ReadyTurn: readyTurn,
	})

}

func (u *UserRecord) BlockInput() {
	u.inputBlocked = true
}

func (u *UserRecord) UnblockInput() {
	u.inputBlocked = false
}

func (u *UserRecord) InputBlocked() bool {
	return u.inputBlocked
}

func (u *UserRecord) CommandFlagged(inputTxt string, flagData events.EventFlag, waitSeconds ...float64) {

	readyTurn := util.GetTurnCount()
	if len(waitSeconds) > 0 {
		readyTurn += uint64(float64(configs.GetConfig().SecondsToTurns(1)) * waitSeconds[0])
	}

	if flagData&events.CmdBlockInput == events.CmdBlockInput {
		u.BlockInput()
	}

	events.AddToQueue(events.Input{
		UserId:    u.UserId,
		InputText: inputTxt,
		ReadyTurn: readyTurn,
		Flags:     flagData,
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

func (u *UserRecord) WimpyCheck() {
	if currentWimpy := u.GetConfigOption(`wimpy`); currentWimpy != nil {
		healthPct := int(math.Floor(float64(u.Character.Health) / float64(u.Character.HealthMax.Value) * 100))
		if healthPct < currentWimpy.(int) {
			u.Command(`flee`, -1)
		}
	}
}
