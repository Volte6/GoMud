
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), 'You are hidden by the admin.')
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "You are no longer admin hidden.")
}
