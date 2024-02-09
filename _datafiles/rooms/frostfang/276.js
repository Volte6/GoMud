/*
props:
- nouns: [leaves, glimmer, light, locket]
  requirements:
  - type: notquesttoken
    idstring: 1-return
    rejectionmessage: You don't see that here anymore.
  description: Nestled inside a pile of leaves is some sort of golden object.
- nouns: [gold, golden, golden locket, locket, sophie's locket, object]
  verbs: [get, take, grab]
  requirements:
  - type: questtoken
    idstring: 1-start
    rejectionmessage: You don't see anything worth taking.
  - type: notquesttoken
    idstring: 1-return
    rejectionmessage: You don't see anything worth taking.
  description: It appears to be a <ansi fg="itemname">golden locket</ansi>.
  trigger:
    itemid: 20025
    questtoken: 1-return
    affected: player
    descriptionplayer: You brush aside the leaves and take the <ansi fg="itemname">golden locket</ansi>.
    descriptionroom: <ansi fg="username">%s</ansi> takes a <ansi fg="itemname">golden locket</ansi> from the pile of leaves.
*/

locketAvailableRound = 0;

const glimmer = ["leaves", "glimmer", "light", "locket"];
const locket = ["gold", "golden", "golden locket", "locket", "sophie's locket", "object"];
const verbs = ["get", "take", "grab"];

function onCommand_look(rest, userId, roomId) {

    roundNow = UtilGetRoundNumber();
    if ( roundNow < locketAvailableRound ) {
        return false;
    }

    if ( UserHasQuest(userId, "1-return") ) {
        return false;
    }

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], glimmer);
        if ( matches.exact.length > 0  ) {


            SendUserMessage(userId, "Nestled inside a pile of leaves is some sort of golden object.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> seems to be digging around in the leaves.");   

            return true;
        }

        matches = UtilFindMatchIn(parts[i], locket);
        if ( matches.exact.length > 0  ) {

            SendUserMessage(userId, "It appears to be a <ansi fg=\"itemname\">golden locket</ansi>.");

            return true;
        }
    }

    return false;
}

// Generic Command Handler
function onCommand(cmd, rest, userId, roomId) {

    roundNow = UtilGetRoundNumber();
    if ( roundNow < locketAvailableRound ) {
        return false;
    }

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    if ( UserHasQuest(userId, "1-return") ) {
        return false;
    }

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], locket);
        if ( matches.exact.length > 0  ) {
            
            SendUserMessage(userId, "You brush aside the leaves and take the <ansi fg=\"itemname\">golden locket</ansi>.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> takes a <ansi fg=\"itemname\">golden locket</ansi> from the pile of leaves.");
            
            UserGiveItem(userId, 20025);
            UserGiveQuest(userId, "1-return");

            locketAvailableRound = roundNow + UtilGetMinutesToRounds(15);

            return true;
        }
    }

    return false;
}



