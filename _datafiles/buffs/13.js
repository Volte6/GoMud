
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), "You begin to feel sick.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is looking sickly.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    dmgAmt = Math.abs(Math.abs(actor.AddHealth(UtilDiceRoll(1, 8)*-1)))

    SendUserMessage(actor.UserId(), 'The poison hurts you for <ansi fg="damage">'+String(dmgAmt)+' damage</ansi>!')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' convulses under the effects of a poison.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "he poison wears off.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' looks a bit more normal.', actor.UserId())
}
