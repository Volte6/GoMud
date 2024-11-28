
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">A warm glow surrounds you.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">A warm glow surrounds '+actor.GetCharacterName(true)+ '.</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">Your glowing fades away.</ansi>' )
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">The glow surrounding '+actor.GetCharacterName(true)+ ' fades away.</ansi>', actor.UserId())
}
