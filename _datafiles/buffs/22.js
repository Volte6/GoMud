
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), 'You catch on <ansi fg="red">fire</ansi>!')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' caught on <ansi fg="red">fire</ansi>!', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    dmgAmt = Math.abs(actor.AddHealth(-1*UtilDiceRoll(2, 6)))

    SendUserMessage(actor.UserId(), 'Flames envelop you, causing <ansi fg="damage">'+String(dmgAmt)+' damage</ansi> while you writh in pain!')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is enveloped in <ansi fg="red">flames</ansi>.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "You are no longer on fire.")
    SendRoomMessage(actor.GetRoomId(), 'The healing aura surrounding '+actor.GetCharacterName(true)+' is no longer on fire.', actor.UserId())
}
