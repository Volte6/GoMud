
const verbs = ["roll", "push"];
const nouns = ["boulder", "rock"];

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    matches = UtilFindMatchIn(rest, nouns);
    if ( !matches.found ) {
        return false;
    }

    if ( room.HasMutator('pushed-boulder') ) {
        SendUserMessage(user.UserId(), "The boulder is already pushed aside.");
        return true;
    }

    SendUserMessage(user.UserId(), "You roll the boulder to the side, revealing a pathway!");

    room.AddMutator('pushed-boulder');
        
    return true;
}

