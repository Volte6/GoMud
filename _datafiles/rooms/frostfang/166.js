
function onCommand_vault(rest, user, room) {

    guard_present = false;

    mobs = room.GetMobs();
    for (i = 0; i < mobs.length; i++) {
        if ( (mob = GetMob(mobs[i])) == null ) {
            continue;
        }
        mobName = mob.GetCharacterName();
        if ( mobName.indexOf("guard") !== -1 ) {
            guard_present = true;
            break;
        }
    }

    hidden = user.HasBuffFlag("hidden");

    if (guard_present && !hidden) {
        SendUserMessage(user.UserId(), "A guard blocks you from entering the vault.");
        SendRoomMessage(room.RoomId(), "A guard blocks "+user.GetCharacterName(true)+" from entering the vault.");
        return true;
    }

    return false;
}

