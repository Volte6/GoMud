
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), 'You feel very well rested.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' looks very well rested.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 2))
}
