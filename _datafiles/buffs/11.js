
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "You are touched by the gods.")
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "The gods forget about you.")
}
