
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'A magical healing aura washes over you.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is surrounded by a healing glow.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 10))

    SendUserMessage(actor.UserId(),     'You heal for <ansi fg="healing">'+String(healAmt)+' damage</ansi>!')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is healing from the effects of a heal spell.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'The healing aura fades away.')
    SendRoomMessage(actor.GetRoomId(),  'The healing aura surrounding '+actor.GetCharacterName(true)+' fades away.', actor.UserId())
}
