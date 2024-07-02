
lastSpawnRound = 0;

// If there is no book here, add the book item
function onEnter(user, room) {

    if ( !user.HasQuest("6-return") ) {
        room.RepeatSpawnItem(10, 30);
    }
    
}
