
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">Your body\'s natural healing feels super charged.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' begins to regenerate.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 3))
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">You regenerate for <ansi fg="healing">'+String(healAmt)+' damage</ansi>!</ansi>')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">Your enhanced regeneration goes away.</ansi>')
}
