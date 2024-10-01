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

func Auction(rest string, user *users.UserRecord) (bool, error) {

	if on := user.GetConfigOption(`auction`); on != nil && !on.(bool) {

		user.SendText(
			`Auctions are disabled. See <ansi fg="command">help set</ansi> for learn how to change this.`,
		)

		return true, nil
	}

	currentAuction := auctions.GetCurrentAuction()

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {

		if currentAuction != nil {
			auctionTxt, _ := templates.Process("auctions/auction-update", currentAuction)
			user.SendText(auctionTxt)
		} else {
			user.SendText(`No current auctions. You can auction something, though!`)
		}
		return true, nil
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
		user.SendText(tplTxt)

		return true, nil
	}

	if args[0] == `bid` {

		if currentAuction == nil {
			user.SendText(`There is not an auction to bid on.`)
			return true, nil
		}

		if currentAuction.SellerUserId == user.UserId {
			user.SendText(`You cannot bid on your own auction.`)
			return true, nil
		}

		if currentAuction.HighestBidUserId == user.UserId {
			user.SendText(`You are already the highest bidder.`)
			return true, nil
		}

		if len(args) < 2 {
			user.SendText(`Bid how much?`)
			return true, nil
		}

		minBid := currentAuction.HighestBid + 1
		if minBid == 0 {
			minBid = currentAuction.MinimumBid
		}

		amt, _ := strconv.Atoi(args[1])
		if amt < minBid {
			user.SendText(fmt.Sprintf(`You must bid at least <ansi fg="gold">%d gold</ansi>.`, minBid))
			return true, nil
		}

		if amt > user.Character.Gold {
			user.SendText(`You don't have that much gold.`)
			return true, nil
		}

		if err := auctions.Bid(user.UserId, amt); err != nil {
			user.SendText(err.Error())
			return true, nil
		}

		user.Character.Gold -= amt

		// Broadcast the bid
		auctionTxt, _ := templates.Process("auctions/auction-bid", currentAuction)
		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					user.SendText(auctionTxt)
				}
			}
		}

		return true, nil
	}

	// If there is already an auction happening, abort this attempt.
	if currentAuction != nil {
		user.SendText(`There is already an auction in progress.`)
		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to auction.", rest))
		return true, nil
	}

	cmdPrompt, _ := user.StartPrompt(`auction`, rest)
	questionConfirm := cmdPrompt.Ask(`Auction your `+matchItem.NameComplex()+`?`, []string{`Yes`, `No`})
	if !questionConfirm.Done {
		return true, nil
	}

	if questionConfirm.Response != `Yes` {
		user.SendText(`Aborting auction`)
		user.ClearPrompt()
		return true, nil
	}

	questionAmount := cmdPrompt.Ask(`Auction for how much gold?`, []string{})
	if !questionAmount.Done {
		return true, nil
	}

	amt, _ := strconv.Atoi(questionAmount.Response)
	if amt < 1 {
		user.SendText(`Aborting auction`)
		user.ClearPrompt()
		return true, nil
	}

	user.ClearPrompt()

	user.SendText(fmt.Sprintf("Auctioning your <ansi fg=\"item\">%s</ansi> for <ansi fg=\"gold\">%d gold</ansi>.", matchItem.DisplayName(), amt))

	if auctions.StartAuction(matchItem, user.UserId, amt) {
		user.Character.RemoveItem(matchItem)
	}

	return true, nil
}
