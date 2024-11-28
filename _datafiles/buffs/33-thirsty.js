
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {

    if ( actor.HasBuffFlag("hydrated")  ) {
        actor.RemoveBuff(33)
        return
    }

    SendUserMessage(actor.UserId(), 'You are feeling parched.');
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {

    if ( actor.HasBuffFlag("hydrated")  ) {
        actor.RemoveBuff(33)
        return
    }

    SendUserMessage(actor.UserId(), 'You feel very thirsty!')
}