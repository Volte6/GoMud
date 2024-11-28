
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">You are touched by the gods.</ansi>')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">The gods forget about you.</ansi>')
}
