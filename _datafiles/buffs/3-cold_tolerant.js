
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You feel warm inside. You feel that you could take on even the harshest winter weather.')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your inner warmth subsides.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' shakes off the haze and the glaze in their eyes fades away.', actor.UserId())
}
