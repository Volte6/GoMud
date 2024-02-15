
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), 'A glowing orb materializes and floats around you.')
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "Your glowing orb fades away.")
}
