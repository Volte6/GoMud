package auctions

import (
	"errors"
	"fmt"
	"time"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/users"
)

type AuctionItem struct {
	ItemData          items.Item
	SellerUserId      int
	SellerName        string
	EndTime           time.Time
	MinimumBid        int
	HighestBid        int
	HighestBidUserId  int
	HighestBidderName string
	LastUpdate        time.Time
}

type PastAuctionItem struct {
	ItemName   string
	WinningBid int
	SellerName string
	BuyerName  string
	EndTime    time.Time
}

var (
	ActiveAuction *AuctionItem = nil
	PastAuctions  []PastAuctionItem
)

func (a *AuctionItem) IsEnded() bool {
	return time.Now().After(a.EndTime)
}

func StartAuction(item items.Item, userId int, minimumBid int) bool {

	if ActiveAuction != nil {
		return false
	}

	if u := users.GetByUserId(userId); u != nil {
		ActiveAuction = &AuctionItem{
			ItemData:          item,
			SellerUserId:      userId,
			SellerName:        u.Character.Name,
			EndTime:           time.Now().Add(time.Second * time.Duration(configs.GetConfig().AuctionSeconds)),
			MinimumBid:        minimumBid,
			HighestBid:        0,
			HighestBidUserId:  0,
			HighestBidderName: ``,
		}
	}

	return true
}

func GetCurrentAuction() *AuctionItem {
	return ActiveAuction
}

func Bid(userId int, bid int) error {

	if ActiveAuction == nil {
		return errors.New("There is not an auction to bid on.")
	}

	if ActiveAuction.HighestBidUserId == userId {
		return errors.New("You are already the highest bidder.")
	}

	if bid < ActiveAuction.MinimumBid || bid < ActiveAuction.HighestBid+1 {
		minBid := ActiveAuction.MinimumBid
		if ActiveAuction.HighestBid > 0 {
			minBid = ActiveAuction.HighestBid + 1
		}
		return fmt.Errorf(`The minimum bid is <ansi fg="gold">%d gold</ansi>`, minBid)
	}

	u := users.GetByUserId(userId)
	if u == nil {
		return errors.New("User not found.")
	}

	ActiveAuction.HighestBid = bid
	ActiveAuction.HighestBidUserId = userId
	ActiveAuction.HighestBidderName = u.Character.Name

	return nil
}

func EndAuction() {

	if ActiveAuction == nil {
		return
	}

	if ActiveAuction.HighestBidUserId != 0 {

		PastAuctions = append(PastAuctions, PastAuctionItem{
			ItemName:   ActiveAuction.ItemData.NameComplex(),
			WinningBid: ActiveAuction.HighestBid,
			SellerName: ActiveAuction.SellerName,
			BuyerName:  ActiveAuction.HighestBidderName,
			EndTime:    ActiveAuction.EndTime,
		})

		if len(PastAuctions) > 30 {
			PastAuctions = PastAuctions[1:]
		}

	}

	ActiveAuction = nil

}

func GetAuctionHistory(totalItems int) []PastAuctionItem {

	if totalItems < 1 {
		return []PastAuctionItem{}
	}

	if totalItems > len(PastAuctions) {
		totalItems = len(PastAuctions)
	}

	return PastAuctions[len(PastAuctions)-totalItems : totalItems]
}
