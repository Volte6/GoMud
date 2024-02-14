
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), 'You enter a focused state of rest.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' begins to meditate.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 4))

    SendUserMessage(actor.UserId(), 'You heal for <ansi fg="damage">'+String(healAmt)+' damage</ansi>.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is healing while they meditate.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "Your restful state abides.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is done meditating.', actor.UserId())
}
