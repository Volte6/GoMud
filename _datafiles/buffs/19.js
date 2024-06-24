
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "You're feeling a little drunk!")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is tatered.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You hiccup!')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' hiccups.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "Your vision straightens out.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' looks sober again.', actor.UserId())
}
