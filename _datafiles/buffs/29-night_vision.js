
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">The shadows come to life as your vision enhances.</ansi>')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">Your vision returns to normal.</ansi>')
}

