package hooks

import (
	"fmt"
	"time"

	"github.com/volte6/gomud/internal/auctions"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

//
// Watches the rounds go by
// Performs auction status updates
//

func AuctionUpdate(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	a := auctions.GetCurrentAuction()
	if a == nil {
		return events.Continue
	}

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

					msg := fmt.Sprintf(`<ansi fg="yellow">You have won the auction for the <ansi fg="item">%s</ansi>! It has been added to your backpack.</ansi>`, a.ItemData.DisplayName())
					user.SendText(msg)
				}
			} else {

				msg := fmt.Sprintf(`You won the auction for the <ansi fg="item">%s</ansi> while you were offline.`, a.ItemData.DisplayName())

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

				msg := fmt.Sprintf(`Your auction of the <ansi fg="item">%s</ansi> has ended. The highest bid was made by <ansi fg="username">%s</ansi> for <ansi fg="gold">%d gold</ansi>.`, a.ItemData.DisplayName(), a.HighestBidderName, a.HighestBid)

				if sellerUser := users.GetByUserId(a.SellerUserId); sellerUser != nil {
					sellerUser.Character.Bank += a.HighestBid
					sellerUser.SendText(`<ansi fg="yellow">` + msg + `</ansi>`)
				} else {

					msg := fmt.Sprintf(`Your auction of the <ansi fg="item">%s</ansi> has ended while you were offline. The highest bid was made by <ansi fg="username">%s</ansi> for <ansi fg="gold">%d gold</ansi>.`, a.ItemData.DisplayName(), a.HighestBidderName, a.HighestBid)

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

					msg := fmt.Sprintf(`<ansi fg="yellow">The auction for the <ansi fg="item">%s</ansi> has ended without a winner. It has been returned to you.</ansi>`, a.ItemData.DisplayName())
					user.SendText(msg)
				}
			}

			for _, uid := range users.GetOnlineUserIds() {
				if uid == a.SellerUserId {
					continue
				}
				if u := users.GetByUserId(uid); u != nil {
					auctionOn := u.GetConfigOption(`auction`)
					if auctionOn == nil || auctionOn.(bool) {
						msg := fmt.Sprintf(`<ansi fg="yellow">The auction for the <ansi fg="item">%s</ansi> has ended without a winner. It has been returned to the seller.</ansi>`, a.ItemData.DisplayName())
						u.SendText(msg)
					}
				}
			}

		}

		sellerName := a.SellerName
		buyerName := a.HighestBidderName
		if a.Anonymous {
			sellerName = `(Anonymous)`
			buyerName = `(Anonymous)`
		}

		events.AddToQueue(events.Auction{
			State:           `END`,
			ItemName:        a.ItemData.NameComplex(),
			ItemDescription: a.ItemData.GetSpec().Description,
			SellerName:      sellerName,
			BuyerName:       buyerName,
			BidAmount:       a.HighestBid,
		})

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

		sellerName := a.SellerName
		buyerName := a.HighestBidderName
		if a.Anonymous {
			sellerName = `(Anonymous)`
			buyerName = `(Anonymous)`
		}

		events.AddToQueue(events.Auction{
			State:           `START`,
			ItemName:        a.ItemData.NameComplex(),
			ItemDescription: a.ItemData.GetSpec().Description,
			SellerName:      sellerName,
			BuyerName:       buyerName,
			BidAmount:       a.HighestBid,
		})

	} else if time.Since(a.LastUpdate) > time.Second*time.Duration(configs.GetAuctionsConfig().UpdateSeconds) {

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

		sellerName := a.SellerName
		buyerName := a.HighestBidderName
		if a.Anonymous {
			sellerName = `(Anonymous)`
			buyerName = `(Anonymous)`
		}

		events.AddToQueue(events.Auction{
			State:           `REMINDER`,
			ItemName:        a.ItemData.NameComplex(),
			ItemDescription: a.ItemData.GetSpec().Description,
			SellerName:      sellerName,
			BuyerName:       buyerName,
			BidAmount:       a.HighestBid,
		})

	}

	return events.Continue
}
