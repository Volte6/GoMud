

crateAvailableRound = 0;

const crate = ["crate", "crates", "box"];
const verbs = ["open", "pry"];

function onCommand_look(rest, user, room) {

    roundNow = UtilGetRoundNumber();

    matches = UtilFindMatchIn(rest, crate);
    if ( matches.found ) {
        if (roundNow < crateAvailableRound) {
            SendUserMessage(user.UserId(), "The scattered crates are broken and empty.");
        } else {
            SendUserMessage(user.UserId(), "The scattered crates are mostly broken and empty, save one that is still intact.");
        }
        
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the broken crates.", user.UserId());   
        return true;
    }

    return false;
}

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    matches = UtilFindMatchIn(rest, crate);
    if ( !matches.found ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();
            
    if (roundNow < crateAvailableRound) {
        SendUserMessage(user.UserId(), "There's nothing here but broken, empty crates.");
        return true;
    }

    if ( !user.HasItemId(10012) ) {
        SendUserMessage(user.UserId(), "You'll need some kind of tool to open that.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" is messing around with an in-tact crate.", user.UserId());   
        return true;
    }

    SendUserMessage(user.UserId(), "You pry the box open and remove a glowing crystal from inside.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" pries the box open and removes something that emits a faint glow.", user.UserId());   
    
    user.GiveItem(4);

    crateAvailableRound = roundNow + UtilGetMinutesToRounds(15)

    return true;
}



