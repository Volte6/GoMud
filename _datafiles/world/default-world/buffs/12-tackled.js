
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re on the ground.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is on the ground.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re trying to get up.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is getting up.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re standing again.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is standing again.', actor.UserId())
}
