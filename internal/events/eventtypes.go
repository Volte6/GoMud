package events

import (
	"time"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/stats"
)

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
	Text            string
	IsCommunication bool
	SourceIsMod     bool
	SkipLineRefresh bool
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
	Config      configs.Config
}

func (n NewRound) Type() string { return `NewRound` }

// Each new turn (TurnMs in config.yaml)
type NewTurn struct {
	TurnNumber uint64
	TimeNow    time.Time
	Config     configs.Config
}

func (n NewTurn) Type() string { return `NewTurn` }

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

type DayNightCycle struct {
	IsSunrise bool
	Day       int
	Month     int
	Year      int
	Time      string
}

func (l DayNightCycle) Type() string { return `DayNightCycle` }

type Auction struct {
	State           string // START, REMINDER, BID, END
	ItemName        string
	ItemDescription string
	SellerName      string
	BuyerName       string
	BidAmount       int
}

func (l Auction) Type() string { return `Auction` }
