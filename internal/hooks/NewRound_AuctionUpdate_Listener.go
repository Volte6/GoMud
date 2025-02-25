package hooks

import (
	"fmt"
	"time"

	"github.com/volte6/gomud/internal/auctions"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

//
// Watches the rounds go by
// Performs auction status updates
//

func AuctionUpdate_Listener(e events.Event) bool {

	evt := e.(events.NewRound)

	a := auctions.GetCurrentAuction()
	if a == nil {
		return true
	}

	c := configs.GetConfig()

	if a.IsEnded() {

		auctions.EndAuction()

		a.LastUpdate = evt.TimeNow
		auctionTxt, _ := templates.Process("auctions/auction-end", a)

		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					u.SendText(auctionTxt)
				}
			}
		}

		// Give the item to the winner and let them know
		if a.HighestBidUserId > 0 {

			if user := users.GetByUserId(a.HighestBidUserId); user != nil {
				if user.Character.StoreItem(a.ItemData) {

					events.AddToQueue(events.ItemOwnership{
						UserId: user.UserId,
						Item:   a.ItemData,
						Gained: true,
					})

					msg := fmt.Sprintf(`<ansi fg="yellow">You have won the auction for the <ansi fg="item">%s</ansi>! It has been added to your backpack.</ansi>%s`, a.ItemData.DisplayName(), term.CRLFStr)
					user.SendText(msg)
				}
			} else {

				msg := fmt.Sprintf(`Your won the auction for the <ansi fg="item">%s</ansi> while you were offline.%s`, a.ItemData.DisplayName(), term.CRLFStr)

				users.SearchOfflineUsers(func(u *users.UserRecord) bool {
					if u.UserId == a.HighestBidUserId {
						user = u
						return false
					}
					return true
				})

				if user != nil {
					user.Inbox.Add(
						users.Message{
							FromName: `Auction System`,
							Message:  msg,
							Item:     &a.ItemData,
						},
					)
					users.SaveUser(*user)
				}

			}

			if a.SellerUserId > 0 {

				msg := fmt.Sprintf(`Your auction of the <ansi fg="item">%s</ansi> has ended while you were offline. The highest bid was made by <ansi fg="username">%s</ansi> for <ansi fg="gold">%d gold</ansi>.%s`, a.ItemData.DisplayName(), a.HighestBidderName, a.HighestBid, term.CRLFStr)

				if sellerUser := users.GetByUserId(a.SellerUserId); sellerUser != nil {
					sellerUser.Character.Bank += a.HighestBid
					sellerUser.SendText(`<ansi fg="yellow">` + msg + `</ansi>`)
				} else {

					users.SearchOfflineUsers(func(u *users.UserRecord) bool {
						if u.UserId == a.SellerUserId {
							sellerUser = u
							return false
						}
						return true
					})

					if sellerUser != nil {
						sellerUser.Inbox.Add(
							users.Message{
								FromName: `Auction System`,
								Message:  msg,
								Gold:     a.HighestBid,
								Item:     &a.ItemData,
							},
						)
						users.SaveUser(*sellerUser)
					}

				}
			}

		} else if a.SellerUserId > 0 {
			if user := users.GetByUserId(a.SellerUserId); user != nil {
				if user.Character.StoreItem(a.ItemData) {

					events.AddToQueue(events.ItemOwnership{
						UserId: user.UserId,
						Item:   a.ItemData,
						Gained: true,
					})

					msg := fmt.Sprintf(`<ansi fg="yellow">The auction for the <ansi fg="item">%s</ansi> has ended without a winner. It has been returned to you.</ansi>%s`, a.ItemData.DisplayName(), term.CRLFStr)
					user.SendText(msg)
				}
			}
		}

	} else if a.LastUpdate.IsZero() {

		a.LastUpdate = evt.TimeNow
		auctionTxt, _ := templates.Process("auctions/auction-start", a)

		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					u.SendText(auctionTxt)
				}
			}
		}

	} else if time.Since(a.LastUpdate) > time.Second*time.Duration(c.AuctionUpdateSeconds) {

		a.LastUpdate = evt.TimeNow
		auctionTxt, _ := templates.Process("auctions/auction-update", a)

		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					u.SendText(auctionTxt)
				}
			}
		}

	}

	return true
}
