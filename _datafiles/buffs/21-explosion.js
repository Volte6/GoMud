
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    dmgAmt = Math.abs(actor.AddHealth(-1*(UtilDiceRoll(2, 9)+2)))

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">Fiery shrapnel hits you for <ansi fg="damage">'+String(dmgAmt)+' damage</ansi>!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">Fiery shrapnel hits '+actor.GetCharacterName(true)+'</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    actor.GiveBuff(22) // On fire
}
