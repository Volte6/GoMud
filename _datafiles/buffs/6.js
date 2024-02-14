
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), "You feel like you've been training in 100x gravity.")
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "You no longer feel like you've been training in 100x gravity.")
}
