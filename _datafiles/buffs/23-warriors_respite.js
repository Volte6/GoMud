
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You enter a focused state of rest.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' begins to meditate.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {

    skillLevel = actor.GetSkillLevel("brawling");
    
    maxHealing = 4;
    if (skillLevel == 3) {
        maxHealing = 6;
    } else if (skillLevel >= 4) {
        maxHealing = 8;
    }

    healAmt = actor.AddHealth(UtilDiceRoll(1, maxHealing))

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You heal for <ansi fg="healing">'+String(healAmt)+' hitpoints</ansi>.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is healing while they meditate.</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">Your restful state abides.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is done meditating.</ansi>', actor.UserId())
}
