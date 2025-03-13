package discord

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Volte6/ansitags"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

var (
	lastMonth = -1
)

// Player enters the world event
func HandlePlayerSpawn(e events.Event) bool {
	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		return false
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return false
	}

	connDetails := connections.Get(user.ConnectionId())

	message := fmt.Sprintf(":white_check_mark: **%s** connected", user.Character.Name)

	if connDetails.IsWebsocket() {
		message += ` (via websocket)`
	}

	SendRichMessage(message, Green)

	return true
}

// Player leaves the world event
func HandlePlayerDespawn(e events.Event) bool {
	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		return false
	}

	message := fmt.Sprintf(":x: **%s** disconnected (online %s)", evt.CharacterName, evt.TimeOnline)

	SendRichMessage(message, Grey)

	return true
}

func HandleLogs(e events.Event) bool {
	evt, typeOk := e.(events.Log)
	if !typeOk {
		return false
	}

	if evt.Level != `ERROR` {
		return true
	}

	msgOut := util.StripANSI(fmt.Sprintln(evt.Data[1:]...))

	// Skip script timeout messages
	if strings.Contains(msgOut, `JSVM`) && strings.Contains(msgOut, `script timeout`) {
		return true
	}

	if strings.Contains(msgOut, `Stopping server`) || strings.Contains(msgOut, `Starting server`) {
		msgOut = strings.Replace(msgOut, evt.Level, `**NOTICE**`, 1)
	} else {
		msgOut = strings.Replace(msgOut, evt.Level, `**`+evt.Level+`**`, 1)
	}

	message := fmt.Sprintf(":bangbang: %s", msgOut)

	SendRichMessage(message, Red)

	return true
}

func HandleLevelup(e events.Event) bool {
	evt, typeOk := e.(events.LevelUp)
	if !typeOk {
		return false
	}

	message := fmt.Sprintf(`:crown: **%s** *has gained a level and reached **level %d**!*`, evt.CharacterName, evt.NewLevel)
	if evt.LevelsGained > 1 {
		message = fmt.Sprintf(`:crown: **%s** *has gained **%d levels** and reached **level %d**!*`, evt.CharacterName, evt.LevelsGained, evt.NewLevel)
	}

	SendRichMessage(message, Gold)

	return true
}

func HandleDeath(e events.Event) bool {
	evt, typeOk := e.(events.PlayerDeath)
	if !typeOk {
		return false
	}

	message := fmt.Sprintf(`:skull: **%s** *has **DIED**!*`, evt.CharacterName)
	SendRichMessage(message, DarkOrange)

	return true
}

func HandleBroadcast(e events.Event) bool {
	evt, typeOk := e.(events.Broadcast)
	if !typeOk {
		return false
	}

	if !evt.IsCommunication {
		return true
	}

	textOut := ansitags.Parse(evt.Text, ansitags.StripTags)

	textOut = strings.ReplaceAll(textOut, "\n", "")
	textOut = strings.ReplaceAll(textOut, "\r", "")
	textOut = strings.Replace(textOut, `(broadcast) `, `(broadcast) **`, 1)
	textOut = strings.Replace(textOut, `:`, `:** `, 1)

	message := fmt.Sprintf(`:speech_balloon: %s`, textOut)

	SendRichMessage(message, Purple)

	return true
}

func HandleAuction(e events.Event) bool {
	evt, typeOk := e.(events.Auction)
	if !typeOk {
		return false
	}

	// Don't spam the reminders.
	if evt.State == `REMINDER` {
		return true
	}

	if evt.State == `BID` {

		itemName := ansitags.Parse(evt.ItemName, ansitags.StripTags)

		payload := webHookPayload{
			Embeds: []embed{{
				Color:       Gold,
				Description: fmt.Sprintf(`:moneybag: **%s** has bid on the auction!`, evt.BuyerName),
				Fields: []embedField{
					{
						Name:   `Amount`,
						Value:  strconv.Itoa(evt.BidAmount),
						Inline: false,
					},
					{
						Name:   `Item`,
						Value:  itemName,
						Inline: true,
					},
					{
						Name:   `Description`,
						Value:  evt.ItemDescription,
						Inline: true,
					},
				},
			}},
		}

		SendPayload(payload)

		return true
	}

	if evt.State == `START` {

		itemName := ansitags.Parse(evt.ItemName, ansitags.StripTags)

		payload := webHookPayload{
			Embeds: []embed{{
				Color:       Gold,
				Description: fmt.Sprintf(`:moneybag: **%s** has started a new auction!`, evt.SellerName),
				Fields: []embedField{
					{
						Name:   `Item`,
						Value:  itemName,
						Inline: true,
					},
					{
						Name:   `Description`,
						Value:  evt.ItemDescription,
						Inline: true,
					},
				},
			}},
		}

		SendPayload(payload)

		return true
	}

	if evt.State == `END` {

		itemName := ansitags.Parse(evt.ItemName, ansitags.StripTags)

		auctionWinner := `No Winner`
		highestBid := `No Bids`

		if evt.BidAmount > 0 {
			auctionWinner = evt.BuyerName
			highestBid = strconv.Itoa(evt.BidAmount)
		}

		payload := webHookPayload{
			Embeds: []embed{{
				Color:       Gold,
				Description: `:moneybag: The auction has ended!`,
				Fields: []embedField{
					{
						Name:   `Highest Bid`,
						Value:  highestBid,
						Inline: true,
					},
					{
						Name:   `Winner`,
						Value:  auctionWinner,
						Inline: true,
					},
					{
						Name:   ` `,
						Value:  ` `,
						Inline: false,
					},
					{
						Name:   `Item`,
						Value:  itemName,
						Inline: true,
					},
					{
						Name:   `Description`,
						Value:  evt.ItemDescription,
						Inline: true,
					},
				},
			}},
		}

		SendPayload(payload)

		return true
	}

	return true
}
