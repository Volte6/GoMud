package parties

import (
	"github.com/volte6/mud/mobs"
)

type Party struct {
	LeaderUserId   int
	UserIds        []int
	MobInstanceIds []int
	InviteUserIds  []int
	AutoAttackers  []int
	Position       map[int]string
}

var (
	partyMap = map[int]*Party{} // key is leader user id, value is party
)

func New(userId int) *Party {
	if _, ok := partyMap[userId]; ok {
		return nil
	}
	p := &Party{
		LeaderUserId:   userId,
		UserIds:        []int{userId},
		InviteUserIds:  []int{},
		MobInstanceIds: []int{},
		AutoAttackers:  []int{},
		Position:       map[int]string{},
	}
	partyMap[userId] = p
	return p
}

func Get(userId int) *Party {
	if party, ok := partyMap[userId]; ok {
		return party
	}
	return nil
}

func IsMobPartied(mobInstId int, userId int) bool {
	if p := Get(userId); p != nil {
		for _, id := range p.GetMobs() {
			if id == mobInstId {
				return true
			}
		}
	}
	return false
}

func (p *Party) ChanceToBeTargetted(userId int) int {

	rank := p.GetRank(userId)
	if rank == `front` {
		return 2
	}

	if rank == `back` {
		return 0
	}
	// middle rank
	return 1
}

func (p *Party) GetRank(userId int) string {
	if val, ok := p.Position[userId]; ok {
		return val
	}
	return `middle`
}

func (p *Party) SetRank(userId int, rank string) {
	if rank == `front` || rank == `back` {
		p.Position[userId] = rank
		return
	}

	delete(p.Position, userId)
}

func (p *Party) IsLeader(userId int) bool {
	return p.LeaderUserId == userId
}

func (p *Party) New(userId int) *Party {
	if party, ok := partyMap[userId]; ok {
		return party
	}
	return nil
}

func (p *Party) SetAutoAttack(userId int, on bool) bool {

	if on {
		for _, id := range p.AutoAttackers {
			if id == userId {
				return true
			}
		}
		p.AutoAttackers = append(p.AutoAttackers, userId)
		return false
	}

	for i, id := range p.AutoAttackers {
		if id == userId {
			p.AutoAttackers = append(p.AutoAttackers[:i], p.AutoAttackers[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Party) GetAutoAttackUserIds() []int {
	return append([]int{}, p.AutoAttackers...)
}

func (p *Party) Leave(userId int) bool {
	if p.IsLeader(userId) {
		if len(p.UserIds) == 1 {
			p.Disband()
			return true
		}

		for _, id := range p.UserIds {
			if id != userId {
				p.LeaderUserId = id
				break
			}
		}
	}

	for i, id := range p.UserIds {
		if id == userId {
			p.UserIds = append(p.UserIds[:i], p.UserIds[i+1:]...)
			break
		}
	}

	for i := len(p.MobInstanceIds) - 1; i >= 0; i-- {
		instId := p.InviteUserIds[i]
		m := mobs.GetInstance(instId)
		if m == nil || m.Character.IsCharmed(userId) {
			p.MobInstanceIds = append(p.MobInstanceIds[:i], p.MobInstanceIds[i+1:]...)
		}
	}

	delete(partyMap, userId)

	return true
}

func (p *Party) IsMember(userId int) bool {
	for _, id := range p.UserIds {
		if id == userId {
			return true
		}
	}
	return false
}

func (p *Party) Invited(userId int) bool {
	for _, id := range p.InviteUserIds {
		if id == userId {
			return true
		}
	}
	return false
}

func (p *Party) InvitePlayer(userId int) bool {
	if _, ok := partyMap[userId]; ok {
		return false
	}
	p.InviteUserIds = append(p.InviteUserIds, userId)
	partyMap[userId] = p

	return true
}

func (p *Party) AcceptInvite(userId int) bool {
	if !p.Invited(userId) {
		return false
	}

	p.UserIds = append(p.UserIds, userId)

	for idx, uid := range p.InviteUserIds {
		if uid == userId {
			p.InviteUserIds = append(p.InviteUserIds[:idx], p.InviteUserIds[idx+1:]...)
			break
		}
	}
	return true
}

func (p *Party) DeclineInvite(userId int) bool {
	if !p.Invited(userId) {
		return false
	}

	for idx, uid := range p.InviteUserIds {
		if uid == userId {
			p.InviteUserIds = append(p.InviteUserIds[:idx], p.InviteUserIds[idx+1:]...)
			break
		}
	}

	delete(partyMap, userId)

	return true
}

func (p *Party) Disband() {
	for _, userId := range p.UserIds {
		delete(partyMap, userId)
	}
	for _, userId := range p.InviteUserIds {
		delete(partyMap, userId)
	}
}

func (p *Party) GetMembers() []int {
	return append([]int{}, p.UserIds...)
}

func (p *Party) AddMob(mobInstanceId int) {
	if len(p.MobInstanceIds) > 0 {
		for _, id := range p.GetMobs() {
			if id == mobInstanceId {
				return
			}
		}
	}
	p.MobInstanceIds = append(p.MobInstanceIds, mobInstanceId)
}

func (p *Party) RemoveMob(mobInstanceId int) {
	for i, id := range p.GetMobs() {
		if id == mobInstanceId {
			p.MobInstanceIds = append(p.MobInstanceIds[:i], p.MobInstanceIds[i+1:]...)
			break
		}
	}
}

func (p *Party) GetMobs() []int {
	ret := []int{}
	for i := len(p.MobInstanceIds) - 1; i >= 0; i-- {
		if m := mobs.GetInstance(p.MobInstanceIds[i]); m != nil {
			ret = append(ret, p.MobInstanceIds[i])
			continue
		}
		p.MobInstanceIds = append(p.MobInstanceIds[:i], p.MobInstanceIds[i+1:]...)
	}
	return ret
}
