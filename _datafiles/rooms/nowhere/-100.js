

lastSpawnRound = 0;

// If there is no book here, add the book item
function onEnter(user, room) {

    if ( !user.HasQuest("0-start") ) {
        user.GiveQuest("0-start");

        trainerMob = room.SpawnMob(53);
        trainerMob.CharmSet(user.UserId(), -1, `despawn`) // -1 means permanent charm (doesn't wear off)
    }
    
}


// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if (cmd != "look" && cmd != "read" ) {
        return false;
    }
    
    if ( rest.substr(rest.length - 3) == "map" || rest.substr(rest.length - 4) == "sign" ) {
      
        SendUserMessage(user.UserId(), "You look at the map nailed to the sign.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" looks at the map nailed to the sign.");

        // Load the cached map, or re-generate and cache it if it's not there
        if ( mapSignData == "" ) {
            mapSignData = GetMap(room.RoomId(), "normal", 22, 38, "Map of Frostfang", false, String(room.RoomId())+",×,Here")
        }

        // Send the map to the user.
        SendUserMessage(user.UserId(), mapSignData);

        return true;
    }
    
    return false;
}
