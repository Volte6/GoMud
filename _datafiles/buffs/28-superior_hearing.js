
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">Even the smallest sounds become clear as your hearing is enhanced.</ansi>')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">Your hearing returns to normal.</ansi>')
}

