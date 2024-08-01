
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The potion warms you as you drink it down.')
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    manaAmt = actor.AddMana(UtilDiceRoll(1, 5))

    SendUserMessage(actor.UserId(), 'You recover <ansi fg="mana-100">'+String(manaAmt)+' mana</ansi>!')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is recovery mana from the effects of a potion.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "The mana potions effect runs out.")
}

