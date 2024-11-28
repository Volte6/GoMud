
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You\'ve been hamstrung!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' has been hamstrung!</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You\'re hamstrung!</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is hamstrung!</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">Your leg heals enough that you can fight again.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is no longer hamstrung.</ansi>', actor.UserId())
}
