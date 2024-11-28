
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">A magical healing aura washes over you.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is surrounded by a healing glow.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 10))

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You heal for <ansi fg="healing">'+String(healAmt)+' damage</ansi>!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is healing from the effects of a heal spell.</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">The healing aura fades away.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">The healing aura surrounding '+actor.GetCharacterName(true)+' fades away.</ansi>', actor.UserId())
}
