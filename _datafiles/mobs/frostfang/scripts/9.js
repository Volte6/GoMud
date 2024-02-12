



function onIdle(mob, room) {

    var random = Math.floor(Math.random() * 10);
    switch (random) {
        case 0:
            action = Math.floor(Math.random() * 3);
            if ( action == 0 ) {
                mob.Command("emote flexes his muscles.");
            } else if ( action == 0 ) {
                mob.Command("emote looks at you suspiciously.");
            } else {
                mob.Command("emote examines his sword carefully.");
            }
            return true;
        case 1:
        case 2:
        case 3:
        case 4:
            return true; // does nothing.
        default:
            break;
    }

    return false;
}