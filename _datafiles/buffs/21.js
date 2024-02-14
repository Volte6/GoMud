
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), 'A magical healing aura washes over you.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is surrounded by a healing glow.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    dmgAmt = Math.abs(actor.AddHealth(-1*(UtilDiceRoll(2, 9)+2)))

    SendUserMessage(actor.UserId(), 'Fiery shrapnel hits you for <ansi fg="damage">'+String(dmgAmt)+' damage</ansi>!')
    SendRoomMessage(actor.GetRoomId(), 'Fiery shrapnel hits '+actor.GetCharacterName(true), actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    actor.GiveBuff(22) // On fire
}
