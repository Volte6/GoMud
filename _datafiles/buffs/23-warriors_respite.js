
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You enter a focused state of rest.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' begins to meditate.', actor.UserId())
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

    SendUserMessage(actor.UserId(), 'You heal for <ansi fg="healing">'+String(healAmt)+' hitpoints</ansi>.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is healing while they meditate.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "Your restful state abides.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is done meditating.', actor.UserId())
}
