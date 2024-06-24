



lastSpawnRound = 0;

// If there is no book here, add the book item
function onEnter(user, room) {

    roundNow = UtilGetRoundNumber();

    nextSpawnRound = lastSpawnRound + UtilGetSecondsToRounds(30);
    if ( lastSpawnRound > 0 && roundNow < nextSpawnRound ) {
        return;
    }

    allItems = room.GetItems();

    oarExists = false;
    for ( i=0; i<allItems.length; i++ ) {
        if ( allItems[i].ItemId() == 10016 ) {
            oarExists = true;
            return;
        }
    }

    if ( !oarExists ) {
        room.SpawnItem(10016, false);
        lastSpawnRound = roundNow;
    }
}


