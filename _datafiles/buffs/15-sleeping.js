
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You lay your head down and immediately doze off.</ansi>');
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is getting some rest.</ansi>', actor.UserId());
    actor.SetAdjective("sleeping", true);
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(3, 8));

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">ZZzzz...</ansi>');
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' snores loudly.</ansi>', actor.UserId());
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You wake up!</ansi>');
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' wakes up.</ansi>', actor.UserId());
    actor.SetAdjective("sleeping", false);
    actor.GiveBuff(16) // Well Rested
}
