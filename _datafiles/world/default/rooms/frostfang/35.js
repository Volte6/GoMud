

const magic_phrase = "zyphrial lumara vorthos";

function onCommand_west(rest, user, room) {
    if ( !user.HasQuest("3-end") ) {
        SendUserMessage(user.UserId(), ' ');
        SendUserMessage(user.UserId(), '<ansi fg="51">The icy wind howls through the gate, and you feel a chill run down your spine.</ansi>');
        SendUserMessage(user.UserId(), '<ansi fg="51">You sense that you are not yet ready to face the dangers that lie beyond.</ansi>');
        SendUserMessage(user.UserId(), ' ');
        
        // Queue it with an input blocking flag and ignore further scripts flag
        user.CommandFlagged('west', EventFlags.CmdSkipScripts|EventFlags.CmdBlockInputUntilComplete, 1)
        // return true (handled) to prevent further execution
        return true

    } 

    SendUserMessage(user.UserId(), '');
    SendUserMessage(user.UserId(), 'The eyes of the stone statues <ansi fg="51">glow</ansi> as you say the words, "<ansi fg="51">'+magic_phrase+'</ansi>"');
    SendUserMessage(user.UserId(), 'You feel a sense of warmth wash over you, and the biting cold air no longer bothers you.');
    SendUserMessage(user.UserId(), '');

    user.GiveBuff(3);

    // Queue it with an input blocking flag and ignore further scripts flag
    user.CommandFlagged('west', EventFlags.CmdSkipScripts|EventFlags.CmdBlockInputUntilComplete, 1)
    // return true (handled) to prevent further execution

    return true
}


function onCommand_say(rest, user, room) {
    
    if ( rest.toLowerCase() !== magic_phrase ) {
        return false;
    }

    SendUserMessage(user.UserId(), ' ');
    SendUserMessage(user.UserId(), 'The eyes of the stone statues <ansi fg="51">glow</ansi> as you say the words aloud.');
    SendUserMessage(user.UserId(), 'You feel a sense of warmth wash over you, and the biting cold air no longer bothers you.');
    SendUserMessage(user.UserId(), ' ');
    
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" chants something unintelligible, and the eyes of the stone statues glow briefly before fading back to ordinary stone.", user.UserId());

    if ( user.HasQuest("3-end") ) {
        return true;
    }

    user.GiveQuest("3-end");

    user.GiveBuff(3);

    return true;
}

