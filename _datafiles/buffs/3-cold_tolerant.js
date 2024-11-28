
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">You feel warm inside. You feel that you could take on even the harshest winter weather.</ansi>')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">Your inner warmth subsides.</ansi>')
    SendRoomMessage(actor.GetRoomId(), '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' shakes off the haze and the glaze in their eyes fades away.</ansi>', actor.UserId())
}
