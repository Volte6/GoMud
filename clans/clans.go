package clans

import (
	"time"

	"github.com/volte6/gomud/items"
)

type ClanRank string

const (
	ClanRankMember     ClanRank = `member`     // normal members get no special privileges
	ClanRankLieutenant ClanRank = `lieutenant` // Lieutenants can accept applications
	ClanRankLeader     ClanRank = `leader`     // Leaders can invite, kick, accept applications and promote members
)

type ClanInfo struct {
	Zone         string       `json:"zone"`         // Zone the clan controls such as "frostfang" or "mystarion"
	ClanTag      string       `json:"clantag"`      // Abbreviated clan name such as "QC", up to 4 characters
	ClanName     string       `json:"clanname"`     // Full clan name such as "Questing Cajuns"
	Upkeep       int          `json:"upkeep"`       // Daily cost in gold to keep the clan going, or it automatically disbands
	MemberUpkeep int          `json:"memberupkeep"` // Daily Gold upkeep cost per member
	Members      []ClanMember `json:"members"`      // List of clan members
	Applications []ClanMember `json:"applications"` // List of clan applications
	Donations    []ClanMember `json:"donations"`    // List of clan donations
}

type ClanMember struct {
	UserId        int       `json:"userid"`        // User ID of the clan member
	CharacterName string    `json:"charactername"` // Character name of the clan member
	Joined        time.Time `json:"joined"`        // Date and time the clan member joined the clan
	Rank          ClanRank  `json:"rank"`          // Rank of the clan member
}

type Donation struct {
	UserId int        `json:"userid"` // User ID of the clan member
	Gold   int        `json:"gold"`   // Amount of gold donated
	Item   items.Item `json:"item"`   // Item donated
	Date   time.Time  `json:"date"`   // Date and time the donation was made
}
