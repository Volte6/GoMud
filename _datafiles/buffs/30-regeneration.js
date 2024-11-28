
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'Your body\'s natural healing feels super charged.')
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' begins to regenerate.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 3))
    SendUserMessage(actor.UserId(), 'You regenerate for <ansi fg="healing">'+String(healAmt)+' damage</ansi>!')
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your enhanced regeneration goes away.')
}
