package usercommands

import (
	"fmt"
	"sort"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

type TrainingOption struct {
	Name          string
	CurrentStatus string
	Cost          int
	Message       string
}
type TrainingOptions struct {
	TrainingPoints int
	Options        []TrainingOption
}

func Train(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if len(room.SkillTraining) == 0 {
		response.SendUserMessage(userId, `You must find a trainer to perform training.`, true)
		return response, nil
	}

	trainingData := TrainingOptions{
		TrainingPoints: user.Character.TrainingPoints,
		Options:        []TrainingOption{}, // name to option
	}

	trainables := []string{}
	maxed := []string{}
	untrainables := []string{}

	trainOpts := map[string]TrainingOption{}

	for skillName, trainingRange := range room.SkillTraining {
		currentLevel := user.Character.GetSkillLevel(skills.SkillTag(skillName))
		opt := TrainingOption{
			Name: skillName,
		}

		if currentLevel >= 4 {
			opt.CurrentStatus = "Maximum"
			opt.Message = ""
			opt.Cost = 0
		} else if currentLevel >= trainingRange.Max {
			opt.CurrentStatus = fmt.Sprintf("Level %d", currentLevel)
			opt.Message = "Seek additional training elsewhere."
			opt.Cost = 0
		} else if currentLevel < trainingRange.Min-1 {
			opt.CurrentStatus = fmt.Sprintf("Level %d", currentLevel)
			opt.Message = fmt.Sprintf("You aren't ready yet. See me for level %d.", trainingRange.Min)
			opt.Cost = 0
		} else if currentLevel >= trainingRange.Min-1 && currentLevel < trainingRange.Max {
			opt.Cost = user.Character.GetSkillLevelCost(currentLevel + 1)
			if currentLevel == 0 {
				opt.CurrentStatus = `<ansi fg="cyan-bold">Unknown</ansi>`
				opt.Message = fmt.Sprintf(`<ansi fg="cyan-bold">Learn this skill here for </ansi><ansi fg="yellow">%d training points</ansi><ansi fg="cyan-bold">.</ansi>`, opt.Cost)
			} else {
				opt.CurrentStatus = fmt.Sprintf(`<ansi fg="cyan-bold">Level %d</ansi>`, currentLevel)
				opt.Message = fmt.Sprintf(`<ansi fg="cyan-bold">Upgrade this skill here for </ansi><ansi fg="yellow">%d training points</ansi><ansi fg="cyan-bold">.</ansi>`, opt.Cost)
			}

			opt.Cost = currentLevel + 1
		}

		trainOpts[skillName] = opt
		if currentLevel >= 4 {
			maxed = append(maxed, skillName)
		} else if opt.Cost > 0 {
			trainables = append(trainables, skillName)
		} else {
			untrainables = append(untrainables, skillName)
		}
	}

	sort.Strings(maxed)
	sort.Strings(trainables)
	sort.Strings(untrainables)

	for _, skillName := range trainables {
		trainingData.Options = append(trainingData.Options, trainOpts[skillName])
	}
	for _, skillName := range maxed {
		trainingData.Options = append(trainingData.Options, trainOpts[skillName])
	}
	for _, skillName := range untrainables {
		trainingData.Options = append(trainingData.Options, trainOpts[skillName])
	}

	if rest == "" {
		exitTxt, _ := templates.Process("descriptions/train", trainingData)
		response.SendUserMessage(userId, exitTxt, false)
	} else {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		allSkills := []string{}
		for skillName, _ := range room.SkillTraining {
			allSkills = append(allSkills, skillName)
		}

		match, closeMatch := util.FindMatchIn(rest, allSkills...)
		if match == "" {
			match = closeMatch
		}

		trainingRange, ok := room.SkillTraining[match]

		currentLevel := user.Character.GetSkillLevel(skills.SkillTag(match))

		if !ok { // If it's not something that can be learned here
			response.SendUserMessage(userId, `The trainer pokes you on your chest, "I think you're in the wrong place, pal."`, true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> looks a little confused.`, user.Character.Name),
				true)
		} else if currentLevel == 4 { // Max level
			response.SendUserMessage(userId, `The trainer chuckles, "I admire your ambition, but you have already mastered that skill!"`, true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`The trainer chuckles and says something you can't quite make out to <ansi fg="username">%s</ansi>`, user.Character.Name),
				true)
		} else if currentLevel < trainingRange.Min-1 { // Not high enough level
			response.SendUserMessage(userId, `The trainer shakes his head, "You aren't ready to train here."`, true)
		} else {

			requiredTrainingPoints := user.Character.GetSkillLevelCost(currentLevel + 1)

			if user.Character.TrainingPoints < requiredTrainingPoints {
				response.SendUserMessage(userId, `The trainer pulls you close and says quietly, "You aren't ready yet. Return when you have more experience."`, true)
				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`The trainer pulls <ansi fg="username">%s</ansi> close and mumbles something in their ear.`, user.Character.Name),
					true)
			} else {

				// Take away the cost
				user.Character.TrainingPoints -= requiredTrainingPoints
				// Upgrade their skill level
				skillName := match
				newLevel := user.Character.TrainSkill(match)

				skillData := struct {
					SkillName  string
					SkillLevel int
				}{
					SkillName:  skillName,
					SkillLevel: newLevel,
				}

				skillUpTxt, _ := templates.Process("character/skillup", skillData)

				response.SendUserMessage(userId, "The trainer grimly considers you for a moment, and then his demeanor changes dramatically.", true)
				response.SendUserMessage(user.UserId, skillUpTxt, true)
				response.SendUserMessage(userId, `"Congratulations!", the trainer exclaims. You are now a little more prepared for the world.`, true)
				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`The trainer shakes <ansi fg="username">%ss</ansi> hand while congratulating them. Must be nice.`, user.Character.Name),
					true)

				if match == string(skills.Tame) {
					if newLevel == 1 {
						user.Character.SetTameCreatureSkill(userId, `rat`, 0)
						response.SendUserMessage(user.UserId, `You've learned how to tame a <ansi fg="mobname">rat</ansi>!`, true)
					}
				}
			}

		}

	}

	response.Handled = true
	return response, nil
}
