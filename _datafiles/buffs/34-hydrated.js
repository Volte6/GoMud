
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {

    actor.CancelBuffWithFlag("thirsty");

    SendUserMessage(actor.UserId(), '<ansi fg="buff-text">Ahhhhhh, life giving water. Nectar of the gods!</ansi>');
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {

    actor.CancelBuffWithFlag("thirsty");
    
}