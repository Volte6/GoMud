
const runes = ["rune", "runes"];
const statues = ["statue", "statues"];
const statue_left = ["zyphrial", "left", "first"];
const statue_right = ["vorthos", "right", "second"];
const magic_phrase = "zyphrial lumara vorthos";

function onCommand_west(rest, userId, roomId) {
    if ( !UserHasQuest(userId, "3-end") ) {
        SendUserMessage(userId, "The icy wind howls through the gate, and you feel a chill run down your spine. You sense that you are not yet ready to face the dangers that lie beyond.");
        return true;
    }
    return false;
}

// cmd specific handler
function onCommand_look(rest, userId, roomId) {

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        runesMatch = UtilFindMatchIn(parts[i], runes);
        if ( runesMatch.exact.length > 0 ) {
            SendUserMessage(userId, "The runes on the gate aren't just decorative; they appear to be part of an old language, possibly used for protective spells or rituals. The runes are only partially readable, and two of the words are scratched out. All you can make out is \"<ansi fg=\"109\">Z-p---l lumara -ort--s.</ansi>\"");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the runes.");
            return true;
        }
    }

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        statuesMatch = UtilFindMatchIn(parts[i], statues);
        if ( statuesMatch.exact.length > 0 ) {
            SendUserMessage(userId, "The statues of the guardian beasts stand as imposing monoliths on either side of the West Gate. Carved from a deep-gray, almost black stone, they depict creatures that seem to be a blend of myth and reality. The one on the left is known as Zyphrial, and on the right is Vorthos.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the stone statues.");
            return true;
        }
    }

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        leftStatueMatch = UtilFindMatchIn(parts[i], statue_left);
        if ( leftStatueMatch.exact.length > 0 ) {
            SendUserMessage(userId, "The statue on the left, named Zyphrial, has the body of a lion, muscular and poised, but its head is that of a majestic eagle with sharp, piercing eyes and a beak that looks ready to snap. Its wings, though folded, span wide, hinting at the power they hold when unfurled.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the statue on the left.");
            return true;
        }
    }

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        leftStatueMatch = UtilFindMatchIn(parts[i], statue_right);
        if ( leftStatueMatch.exact.length > 0 ) {
            SendUserMessage(userId, "The statue on the right, known as Vorthos, is serpentine, its long, coiled body reminiscent of a dragon. It has the scales of a reptile, but its face is almost humanoid, with deep-set eyes and a wise, contemplative expression. Twin horns spiral upwards from its forehead, and its clawed feet grip the base as if it's ready to pounce.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the statue on the right.");
            return true;
        }
    }


    return false;
}

function onCommand_say(rest, userId, roomId) {
    
    if ( rest.toLowerCase() !== magic_phrase ) {
        return false;
    }

    SendUserMessage(userId, "The eyes of the stone statues glow as you say the words aloud. You feel a sense of warmth wash over you, and the biting cold air no longer bothers you.");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> chants something unintelligible, and the eyes of the stone statues glow briefly before fading back to ordinary stone.");

    if ( UserHasQuest(userId, "3-end") ) {
        console.log("User has quest 3-end");
        return true;
    }

    UserGiveQuest(userId, "3-end");

    return true;
}