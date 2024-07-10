
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'A light materializes and engulfs you.')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "Your glowing fades away.")
}
