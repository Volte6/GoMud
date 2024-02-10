
mapSignData = ""

// Generic Command Handler
function onCommand(cmd, rest, userId, roomId) {

    if (cmd != "look" && cmd != "read" ) {
        return false;
    }
    if ( rest.substr(rest.length - 3) == "map" || rest.substr(rest.length - 4) == "sign" ) {
      
        SendUserMessage(userId, "You look at the map nailed to the sign.");
        SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> looks at the map nailed to the sign.");

        // Load the cached map, or re-generate and cache it if it's not there
        if ( mapSignData == "" ) {
            mapSignData = RoomGetMap(roomId, "normal", 22, 38, "Map of Frostfang", false, String(roomId)+",×,Here")
        }

        // Send the map to the user.
        SendUserMessage(userId, mapSignData);

        return true;
    }
    
    return false;
}

// Executes when the room first loads.
function onLoad(roomId) {
    // Just running this to pre-cache the map so that if someone looks at the map it won't time out
    mapSignData = RoomGetMap(roomId, "normal", 22, 38, "Map of Frostfang", false, String(roomId)+",×,Here")
}

// return true - allow to enter
// return false - leave them in the room the started
function onEnter(userId, roomId) {
    reject = false;
    SendUserMessage(userId, "You walk into the Town Square.")
    return reject;
}

// return true - proces exit normally
// return false - disallow the exit and keep them in their current room
function onExit(userId, roomId) {
    reject = false;
    SendUserMessage(userId, "You leave the Town Square.")
    return reject;
}
