package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Pet(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		user.SendText(`Pet what?`)
		return true, nil
	}

	if args[0] == `name` {

		if !user.Character.Pet.Exists() {
			user.SendText(`You have no pet to name.`)
			return true, nil
		}

		if user.Character.Pet.Name != `` && user.Character.Pet.Name != user.Character.Pet.Type {
			user.SendText(fmt.Sprintf(`%s already has a name.`, user.Character.Pet.DisplayName()))
			return true, nil
		}

		newName := strings.Join(args[1:], ` `)

		if err := util.ValidateName(newName); err != nil {
			user.SendText(`That name is not allowed: ` + err.Error())
			return true, nil
		}

		if configs.GetConfig().IsBannedName(newName) {
			user.SendText(`That name is prohibited.`)
			return true, nil
		}

		for _, name := range mobs.GetAllMobNames() {
			if strings.EqualFold(name, newName) {
				user.SendText(`That name is prohibited.`)
				return true, nil
			}
		}

		user.Character.Pet.Name = newName

		user.SendText(fmt.Sprintf(`You name your pet: %s.`, user.Character.Pet.DisplayName()))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> names their pet %s`, user.Character.Name, user.Character.Pet.DisplayName()), user.UserId)

		// rename their pet?
		return true, nil
	}

	// Map name to display name
	petDisplayNames := map[string]string{}
	petNames := []string{}

	if user.Character.Pet.Exists() {
		petDisplayNames[user.Character.Pet.Name] = user.Character.Pet.DisplayName()
		petNames = append(petNames, user.Character.Pet.Name)
	}

	for _, uId := range room.GetPlayers() {
		if uId == user.UserId {
			continue
		}

		if u := users.GetByUserId(uId); u != nil {
			if u.Character.Pet.Exists() {
				petDisplayNames[u.Character.Pet.Name] = u.Character.Pet.DisplayName()
				petNames = append(petNames, u.Character.Pet.Name)
			}
		}
	}

	match, closeMatch := util.FindMatchIn(rest, petNames...)
	if match == `` {
		match = closeMatch
	}

	if match == `` {
		user.SendText(`Can't find that to pet.`)
		return true, nil
	}

	user.SendText(fmt.Sprintf(`You pet %s`, petDisplayNames[match]))

	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> pets %s`, user.Character.Name, petDisplayNames[match]), user.UserId)

	roll := util.RollDice(1, 4)

	if roll == 1 {
		room.SendText(fmt.Sprintf(`%s twirls a bit.`, petDisplayNames[match]))
	}

	if roll == 2 {
		room.SendText(fmt.Sprintf(`%s stiffens.`, petDisplayNames[match]))
	}

	return true, nil
}
