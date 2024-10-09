package pets

type Power string

const (
	Attack         Power = `attack`
	SeeHidden      Power = `see-hidden`
	CarryItems     Power = `carry-items`      // Carry 5 items
	CarryItemsMore Power = `carry-items-more` // Carry 10 items
	SeeNouns       Power = `see-nouns`
)
