
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re exhausted!')
    SendRoomMessage(actor.GetRoomId(),  '' + actor.GetCharacterName(true)+' is exhausted.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re exhausted!')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is exhausted!', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You catch your breath.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is no longer exhausted.', actor.UserId())
}
