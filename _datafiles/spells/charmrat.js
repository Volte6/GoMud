

// Called when the casting is initialized (cast command)
// Return false if the casting should be ignored/aborted
function onCast(sourceActor, targetActor) {

    race = targetActor.GetRace();

    if ( race.toLowerCase() != "rodent" ) {
        SendUserMessage(sourceActor.UserId(), 'This spell only works on rodents.');
        return false;
    }

    SendUserMessage(sourceActor.UserId(), 'You attempt to beguile the '+targetActor.GetCharacterName(true)+'.');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' performs rhythmic chanting directed at '+targetActor.GetCharacterName(true)+'.', sourceActor.UserId());

    return true
}

function onWait(sourceActor, targetActor) {

    SendUserMessage(sourceActor.UserId(), 'Your chanting and gyrating intensifies.');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' continues chanting and gyrating...', sourceActor.UserId());
}

// Called when the spell succeeds its cast attempt
// Return true to ignore any auto-retaliation from the target
function onMagic(sourceActor, targetActor) {

    smarts = sourceActor.GetStat("smarts") - targetActor.GetStat("smarts");
    mys = sourceActor.GetStat("mysticism") - targetActor.GetStat("mysticism");

    charmRounds = Math.ceil((smarts + mys) / 2)

    roomId = sourceActor.GetRoomId();

    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);

    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    targetActor.CharmSet(sourceActor.UserId(), charmRounds);

    // Tell the caster about the action
    SendUserMessage(sourceUserId, 'The '+targetName+' has been charmed by you for '+String(charmRounds)+' rounds!');

    // Tell the room about the heal, except the source and target
    SendRoomMessage(roomId, sourceActor.GetCharacterName(true)+' charms the '+targetName+'!', sourceUserId, targetUserId);

    // Tell the target about the heal
    if ( targetUserId != 0 ) {
        SendUserMessage(targetUserId, sourceName+' has charmed you!');
    }
 
    return true;
}

