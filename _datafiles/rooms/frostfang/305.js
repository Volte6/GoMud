

crateAvailableRound = 0;

const caravan = ["caravan", "caravans", "wagon", "wagons"];
const crate = ["crate", "crates", "box"];
const verbs = ["open", "pry"];

function onCommand_look(rest, user, room) {

    roundNow = UtilGetRoundNumber();

    matches = UtilFindMatchIn(rest, caravan);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "The caravan, long since destroyed, once belonged to the frostfire guild of magicians. There must have been some impressive artifacts carreid by them once.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the caravan.");   
        return true;
    }

    matches = UtilFindMatchIn(rest, crate);
    if ( matches.found ) {
        if (roundNow < crateAvailableRound) {
            SendUserMessage(user.UserId(), "The scattered crates are broken and empty.");
        } else {
            SendUserMessage(user.UserId(), "The scattered crates are mostly broken and empty, save one that is still intact.");
        }
        
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the broken crates.");   
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
        SendUserMessage(user.UserId(), "There's nothign here but broken, empty crates.");
        return true;
    }

    if ( !user.HasItemId(10012) ) {
        SendUserMessage(user.UserId(), "You'll need some kind of tool to open that.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" is messing around with an in-tact crate.");   
        return true;
    }

    SendUserMessage(user.UserId(), "You pry the box open and remove a glowing crystal from inside.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" pries the box open and removes something that emits a faint glow.");   
    
    user.GiveItem(4);

    crateAvailableRound = roundNow + UtilGetMinutesToRounds(15)

    return true;
}



