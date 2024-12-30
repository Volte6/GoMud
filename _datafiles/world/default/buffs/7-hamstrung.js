
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'ve been hamstrung!')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' has been hamstrung!', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re hamstrung!')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is hamstrung!', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'Your leg heals enough that you can fight again.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is no longer hamstrung.', actor.UserId())
}
