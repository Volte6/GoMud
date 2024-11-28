
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You begin to feel sick.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is looking sickly.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    dmgAmt = Math.abs(Math.abs(actor.AddHealth(UtilDiceRoll(1, 8)*-1)))

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">The poison hurts you for <ansi fg="damage">'+String(dmgAmt)+' damage</ansi>!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' convulses under the effects of a poison.</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">The poison wears off.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' looks a bit more normal.</ansi>', actor.UserId())
}
