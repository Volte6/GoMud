
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The shadows come to life as your vision enhances.')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "Your vision returns to normal.")
}

