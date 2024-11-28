
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You feel very well rested.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' looks very well rested.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 2))
}
