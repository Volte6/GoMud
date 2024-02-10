package characters

const (
	CharmPermanent      = -1 // Never expires due to time
	CharmExpiredDespawn = `emote bows and bids you farewell, disappearing into the scenery;despawn charmed mob expired`
	CharmExpiredRevert  = `emote reverts to its old ways`
)

var ()

type CharmInfo struct {
	UserId          int    // Charmed or serving a player?
	RoundsRemaining int    // If -2, never expires
	ExpiredCommand  string // Any valid mob commands such as `emote bows and waves farewell;despawn`
}

func NewCharm(userId int, rounds int, expireCommand string) *CharmInfo {
	return &CharmInfo{UserId: userId, RoundsRemaining: rounds, ExpiredCommand: expireCommand}
}

func (ch *CharmInfo) Expire() {
	ch.RoundsRemaining = 0
}
