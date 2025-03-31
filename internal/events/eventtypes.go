package events

import (
	"strconv"
	"time"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/stats"
)

// EVENT DEFINITIONS FOLLOW
// NOTE: If you give an event the following receiver function: `UniqueID() string`
//
//	      It will become a "unique event", meaning only one can be in the event queue
//			 at a time matching the string return value.
//		 Example: See `RedrawPrompt`
//
// Used to apply or remove buffs
type Buff struct {
	UserId        int
	MobInstanceId int
	BuffId        int
}

func (b Buff) Type() string { return `Buff` }

// Used for giving/taking quest progress
type Quest struct {
	UserId     int
	QuestToken string
}

func (q Quest) Type() string { return `Quest` }

// For special room-targetting actions
type RoomAction struct {
	RoomId       int
	SourceUserId int
	SourceMobId  int
	Action       string
	Details      any
	ReadyTurn    uint64
}

func (r RoomAction) Type() string { return `RoomAction` }

// Used for Input from players/mobs
type Input struct {
	UserId        int
	MobInstanceId int
	InputText     string
	ReadyTurn     uint64
	Flags         EventFlag
}

func (i Input) Type() string { return `Input` }

// Messages that are intended to reach all users on the system
type Broadcast struct {
	Text             string
	TextScreenReader string // optional text for screenreader friendliness
	IsCommunication  bool
	SourceIsMod      bool
	SkipLineRefresh  bool
}

func (b Broadcast) Type() string { return `Broadcast` }

type Message struct {
	UserId          int
	ExcludeUserIds  []int
	RoomId          int
	Text            string
	IsQuiet         bool // whether it can only be heard by superior "hearing"
	IsCommunication bool // If true, this is a communication such as "say" or "emote"
}

func (m Message) Type() string { return `Message` }

// Special commands that only the webclient is equipped to handle
type WebClientCommand struct {
	ConnectionId uint64
	Text         string
}

func (w WebClientCommand) Type() string { return `WebClientCommand` }

// GMCP Commands from clients to server
type GMCPIn struct {
	ConnectionId uint64
	Command      string
	Json         []byte
}

func (g GMCPIn) Type() string { return `GMCPIn` }

// GMCP Commands from server to client
type GMCPOut struct {
	UserId  int
	Module  string
	Payload any
}

func (g GMCPOut) Type() string { return `GMCPOut` }

// Messages that are intended to reach all users on the system
type System struct {
	Command string
	Data    any
}

func (s System) Type() string { return `System` }

// Payloads describing sound/music to play
type MSP struct {
	UserId    int
	SoundType string // SOUND or MUSIC
	SoundFile string
	Volume    int    // 1-100
	Category  string // special category/type for MSP string
}

func (m MSP) Type() string { return `MSP` }

// Fired whenever a mob or player changes rooms
type RoomChange struct {
	UserId        int
	MobInstanceId int
	FromRoomId    int
	ToRoomId      int
}

func (r RoomChange) Type() string { return `RoomChange` }

// Fired every new round
type NewRound struct {
	RoundNumber uint64
	TimeNow     time.Time
}

func (n NewRound) Type() string { return `NewRound` }

// Each new turn (TurnMs in config.yaml)
type NewTurn struct {
	TurnNumber uint64
	TimeNow    time.Time
}

func (n NewTurn) Type() string { return `NewTurn` }

// Gained or lost an item
type EquipmentChange struct {
	UserId        int
	MobInstanceId int
	GoldChange    int
	BankChange    int
	ItemsWorn     []items.Item
	ItemsRemoved  []items.Item
}

func (i EquipmentChange) Type() string { return `EquipmentChange` }

// Gained or lost an item
type ItemOwnership struct {
	UserId        int
	MobInstanceId int
	Item          items.Item
	Gained        bool
}

func (i ItemOwnership) Type() string { return `ItemOwnership` }

// Triggered by a script
type ScriptedEvent struct {
	Name string
	Data map[string]any
}

func (s ScriptedEvent) Type() string { return `ScriptedEvent` }

// Entered the world
type PlayerSpawn struct {
	UserId        int
	RoomId        int
	Username      string
	CharacterName string
}

func (p PlayerSpawn) Type() string { return `PlayerSpawn` }

// Left the world
type PlayerDespawn struct {
	UserId        int
	RoomId        int
	Username      string
	CharacterName string
	TimeOnline    string
}

func (p PlayerDespawn) Type() string { return `PlayerDespawn` }

type Log struct {
	FollowAdd    connections.ConnectionId
	FollowRemove connections.ConnectionId
	Level        string
	Data         []any
}

func (l Log) Type() string { return `Log` }

type LevelUp struct {
	UserId         int
	RoomId         int
	Username       string
	CharacterName  string
	LevelsGained   int
	NewLevel       int
	StatsDelta     stats.Statistics
	TrainingPoints int
	StatPoints     int
	LivesGained    int
}

func (l LevelUp) Type() string { return `LevelUp` }

type PlayerDeath struct {
	UserId        int
	RoomId        int
	Username      string
	CharacterName string
	Permanent     bool
	KilledByUsers []int
}

func (l PlayerDeath) Type() string { return `PlayerDeath` }

type MobDeath struct {
	MobId         int
	InstanceId    int
	RoomId        int
	CharacterName string
	Level         int
	PlayerDamage  map[int]int
}

func (l MobDeath) Type() string { return `MobDeath` }

type DayNightCycle struct {
	IsSunrise bool
	Day       int
	Month     int
	Year      int
	Time      string
}

func (l DayNightCycle) Type() string { return `DayNightCycle` }

type Looking struct {
	UserId int
	RoomId int
	Target string
	Hidden bool
}

func (l Looking) Type() string { return `Looking` }

// Fired after creating a new character and giving the character a name.
type CharacterCreated struct {
	UserId        int
	CharacterName string
}

func (p CharacterCreated) Type() string { return `CharacterCreated` }

// Fired when a character alt change has occured.
type CharacterChanged struct {
	UserId            int
	LastCharacterName string
	CharacterName     string
}

func (p CharacterChanged) Type() string { return `CharacterChanged` }

type UserSettingChanged struct {
	UserId int
	Name   string
}

func (i UserSettingChanged) Type() string { return `UserSettingChanged` }

// Health, mana, etc.
type CharacterVitalsChanged struct {
	UserId int
}

func (p CharacterVitalsChanged) Type() string { return `CharacterVitalsChanged` }

// Health, mana, etc.
type CharacterTrained struct {
	UserId int
}

func (p CharacterTrained) Type() string { return `CharacterTrained` }

type RedrawPrompt struct {
	UserId        int
	OnlyIfChanged bool
}

func (l RedrawPrompt) Type() string     { return `RedrawPrompt` }
func (l RedrawPrompt) UniqueID() string { return `RedrawPrompt-` + strconv.Itoa(l.UserId) }
