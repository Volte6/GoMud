
HEAL_DICE_QTY = 1
HEAL_DICE_SIDES = 2
SPELL_DESCRIPTION = '<ansi fg="222">sample helpful single target spell</ansi>'

// Called when the casting is initialized (cast command)
// Return false if the casting should be ignored/aborted
function onCast(sourceActor, targetActor) {

    SendUserMessage(sourceActor.UserId(), 'You begin to chant softly.');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' begins to chant softly.', sourceActor.UserId());
    return true
}

function onWait(sourceActor, targetActor) {

    SendUserMessage(sourceActor.UserId(), 'You continue chanting...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' continues chanting...', sourceActor.UserId());
}

// Called when the spell succeeds its cast attempt
function onMagic(sourceActor, targetActor) {

    roomId = sourceActor.GetRoomId();

    healAmt = UtilDiceRoll(HEAL_DICE_QTY, HEAL_DICE_SIDES);
    healAmtStr = String(healAmt);

    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);

    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    if ( sourceActor.UserId() != targetActor.UserId() ) {

        // Tell the caster about the action
        SendUserMessage(sourceUserId, 'You stop chanting and casts '+SPELL_DESCRIPTION+' on '+targetName+', healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>.');

        // Tell the room about the heal, except the source and target
        SendRoomMessage(roomId, sourceName+' stops chanting and casts '+SPELL_DESCRIPTION+' on '+targetName+', providing health.', sourceUserId, targetUserId);

        // Tell the target about the heal
        SendUserMessage(targetUserId, sourceName+' stops chanting and casts '+SPELL_DESCRIPTION+' on you, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>.');

    } else {

        // Tell the cast they did it to themselves
        SendUserMessage(sourceUserId, 'You stop chanting and cast '+SPELL_DESCRIPTION+' on yourself, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>.');

        // Tell the room about the heal, except the source and target
        SendRoomMessage(roomId, sourceName+' stops chanting and casts '+SPELL_DESCRIPTION+' on themselves, providing health.', sourceUserId, targetUserId);

    }

    // Apply the heal to the target
    targetActor.AddHealth(healAmt);
    
}
