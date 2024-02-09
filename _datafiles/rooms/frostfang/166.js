
function onCommand_vault(rest, userId, roomId) {

    guard_present = false;

    mobs = RoomGetMobs(roomId);
    for (i = 0; i < mobs.length; i++) {
    
        mobName = MobGetCharacterName(mobs[i]);
        if ( mobName.indexOf("guard") !== -1 ) {
            guard_present = true;
            break;
        }
    }

    hidden = UserHasBuffFlag(userId, "hidden");

    if (guard_present && !hidden) {
        SendUserMessage(userId, "A guard blocks you from entering the vault.");
        SendRoomMessage(roomId, "A guard blocks <ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> from entering the vault.");
        return true;
    }

    return false;
}

