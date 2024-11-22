package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
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

		if bannedPattern, ok := configs.GetConfig().IsBannedName(newName); ok {
			user.SendText(`that username matched the prohibited name pattern: "` + bannedPattern + `"`)
			return true, nil
		}

		for _, name := range mobs.GetAllMobNames() {
			if strings.EqualFold(name, newName) {
				user.SendText(`That name is prohibited.`)
				return true, nil
			}
		}

		if foundUserId, _ := users.CharacterNameSearch(newName); foundUserId > 0 {
			user.SendText(`There's a character with that name. That would be too confusing!`)
			return true, nil
		}

		user.Character.Pet.Name = newName

		user.SendText(fmt.Sprintf(`You name your pet: %s.`, user.Character.Pet.DisplayName()))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> names their pet %s`, user.Character.Name, user.Character.Pet.DisplayName()), user.UserId)

		// rename their pet?
		return true, nil
	}

	petUserId := room.FindByPetName(rest)
	if petUserId == 0 && rest == `pet` && user.Character.Pet.Exists() {
		petUserId = user.UserId
	}
	if petUserId == 0 {
		user.SendText(`Can't find that to pet.`)
		return true, nil
	}

	petUser := users.GetByUserId(petUserId)
	if petUser == nil {
		user.SendText(`Can't find that to pet.`)
		return true, nil
	}

	user.SendText(fmt.Sprintf(`You pet %s`, petUser.Character.Pet.DisplayName()))

	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> pets %s`, user.Character.Name, petUser.Character.Pet.DisplayName()), user.UserId)

	roll := util.RollDice(1, 4)

	if roll == 1 {
		room.SendText(fmt.Sprintf(`%s twirls a bit.`, petUser.Character.Pet.DisplayName()))
	} else if roll == 2 {
		room.SendText(fmt.Sprintf(`%s stiffens.`, petUser.Character.Pet.DisplayName()))
	}

	return true, nil
}
