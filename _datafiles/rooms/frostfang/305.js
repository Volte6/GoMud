

crateAvaialbleRound = 0;

const caravan = ["caravan", "caravans", "wagon", "wagons"];
const crate = ["crate", "crates", "box"];
const verbs = ["open", "pry"];

function onCommand_look(rest, userId, roomId) {

    roundNow = UtilGetRoundNumber();

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], caravan);
        if ( matches.exact.length > 0  ) {
            SendUserMessage(userId, "The caravan, long since destroyed, once belonged to the frostfire guild of magicians. There must have been some impressive artifacts carreid by them once.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the caravan.");   
            return true;
        }

        matches = UtilFindMatchIn(parts[i], crate);
        if ( matches.exact.length > 0  ) {

            if (roundNow < crateAvaialbleRound) {
                SendUserMessage(userId, "The scattered crates are broken and empty.");
            } else {
                SendUserMessage(userId, "The scattered crates are mostly broken and empty, save one that is still intact.");
            }
            
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the broken crates.");   
            return true;
        }
    }

    return false;
}

// Generic Command Handler
function onCommand(cmd, rest, userId, roomId) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], crate);
        if ( matches.exact.length > 0  ) {
            
            if (roundNow < crateAvaialbleRound) {
                SendUserMessage(userId, "There's nothign here but broken, empty crates.");
                return true;
            }

            if ( !UserHasItemId(userId, 10012) ) {
                SendUserMessage(userId, "You'll need some kind of tool to open that.");
                SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> is messing around with an in-tact crate.");   
                return true;
            }

            SendUserMessage(userId, "You pry the box open and remove a glowing crystal from inside.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> pries the box open and removes something that emits a faint glow.");   
            UserGiveItem(userId, 4);

            crateAvaialbleRound = roundNow + UtilGetMinutesToRounds(15)

            return true;
        }
    }

    return false;
}



