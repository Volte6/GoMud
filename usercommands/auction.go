package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/auctions"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Auction(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if on := user.GetConfigOption(`auction`); on != nil && !on.(bool) {

		response.SendUserMessage(userId,
			`Auctions are disabled. See <ansi fg="command">help set</ansi> for learn how to change this.`,
			true)

		response.Handled = true
		return response, nil
	}

	currentAuction := auctions.GetCurrentAuction()

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {

		if currentAuction != nil {
			auctionTxt, _ := templates.Process("auctions/auction-update", currentAuction)
			response.SendUserMessage(userId, auctionTxt, true)
		} else {
			response.SendUserMessage(userId, `No current auctions. You can auction something, though!`, true)
		}
		response.Handled = true
		return response, nil
	}

	if args[0] == `history` {

		headers := []string{"Date", "Item", "Seller", "Buyer", "Winning Bid"}
		formatting := []string{
			`<ansi fg="magenta">%s</ansi>`,
			`<ansi fg="item">%s</ansi>`,
			`<ansi fg="username">%s</ansi>`,
			`<ansi fg="username">%s</ansi>`,
			`<ansi fg="gold">%s</ansi>`,
		}

		rows := [][]string{}

		auctionHistory := auctions.GetAuctionHistory(0)

		for i := len(auctionHistory) - 1; i >= 0; i-- {
			aItem := auctionHistory[i]

			buyerName := aItem.BuyerName
			sellerName := aItem.SellerName
			if aItem.Anonymous {
				buyerName = `Anonymous`
				sellerName = `Anonymous`
			}
			rows = append(rows, []string{
				aItem.EndTime.Format("2006-01-02 15:04:05"),
				aItem.ItemName,
				sellerName,
				buyerName,
				strconv.Itoa(aItem.WinningBid) + " gold",
			})
		}

		historyTableData := templates.GetTable(`Past Auctions`, headers, rows, formatting)

		tplTxt, _ := templates.Process("tables/generic", historyTableData)
		response.SendUserMessage(userId, tplTxt, true)

		response.Handled = true
		return response, nil
	}

	if args[0] == `bid` {

		if currentAuction == nil {
			response.SendUserMessage(userId, `There is not an auction to bid on.`, true)
			response.Handled = true
			return response, nil
		}

		if currentAuction.SellerUserId == userId {
			response.SendUserMessage(userId, `You cannot bid on your own auction.`, true)
			response.Handled = true
			return response, nil
		}

		if currentAuction.HighestBidUserId == userId {
			response.SendUserMessage(userId, `You are already the highest bidder.`, true)
			response.Handled = true
			return response, nil
		}

		if len(args) < 2 {
			response.SendUserMessage(userId, `Bid how much?`, true)
			response.Handled = true
			return response, nil
		}

		minBid := currentAuction.HighestBid + 1
		if minBid == 0 {
			minBid = currentAuction.MinimumBid
		}

		amt, _ := strconv.Atoi(args[1])
		if amt < minBid {
			response.SendUserMessage(userId, fmt.Sprintf(`You must bid at least <ansi fg="gold">%d gold</ansi>.`, minBid), true)
			response.Handled = true
			return response, nil
		}

		if amt > user.Character.Gold {
			response.SendUserMessage(userId, `You don't have that much gold.`, true)
			response.Handled = true
			return response, nil
		}

		if err := auctions.Bid(userId, amt); err != nil {
			response.SendUserMessage(userId, err.Error(), true)
			response.Handled = true
			return response, nil
		}

		user.Character.Gold -= amt

		// Broadcast the bid
		auctionTxt, _ := templates.Process("auctions/auction-bid", currentAuction)
		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					response.SendUserMessage(uid, auctionTxt, true)
				}
			}
		}

		response.Handled = true
		return response, nil
	}

	// If there is already an auction happening, abort this attempt.
	if currentAuction != nil {
		response.SendUserMessage(userId, `There is already an auction in progress.`, true)
		response.Handled = true
		return response, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf("You don't have a %s to auction.", rest), true)
		response.Handled = true
		return response, nil
	}

	cmdPrompt, _ := user.StartPrompt(`auction`, rest)
	questionConfirm := cmdPrompt.Ask(`Auction your `+matchItem.NameComplex()+`?`, []string{`Yes`, `No`})
	if !questionConfirm.Done {
		response.Handled = true
		return response, nil
	}

	if questionConfirm.Response != `Yes` {
		response.SendUserMessage(userId, `Aborting auction`, true)
		user.ClearPrompt()
		response.Handled = true
		return response, nil
	}

	questionAmount := cmdPrompt.Ask(`Auction for how much gold?`, []string{})
	if !questionAmount.Done {
		response.Handled = true
		return response, nil
	}

	amt, _ := strconv.Atoi(questionAmount.Response)
	if amt < 1 {
		response.SendUserMessage(userId, `Aborting auction`, true)
		user.ClearPrompt()
		response.Handled = true
		return response, nil
	}

	user.ClearPrompt()

	response.SendUserMessage(userId, fmt.Sprintf("Auctioning your <ansi fg=\"item\">%s</ansi> for <ansi fg=\"gold\">%d gold</ansi>.", matchItem.DisplayName(), amt), true)

	if auctions.StartAuction(matchItem, userId, amt) {
		user.Character.RemoveItem(matchItem)
	}

	response.Handled = true
	return response, nil
}
