
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), "Weakness overtakes your body!")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' looks a little shakey.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    SendUserMessage(actor.UserId(), "You're feeling weak!")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' looks a little shakey.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "You feel a little more like yourself.")
}
