
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You catch on <ansi fg="red">fire</ansi>!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' caught on <ansi fg="red">fire</ansi>!</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    dmgAmt = Math.abs(actor.AddHealth(-1*UtilDiceRoll(2, 6)))

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">Flames envelop you, causing <ansi fg="damage">'+String(dmgAmt)+' damage</ansi> while you writh in pain!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is enveloped in <ansi fg="red">flames</ansi>.</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You are no longer on fire.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">The healing aura surrounding '+actor.GetCharacterName(true)+' is no longer on fire.</ansi>', actor.UserId())
}
