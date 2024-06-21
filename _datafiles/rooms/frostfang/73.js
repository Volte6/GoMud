
lastSpawnRound = 0;

// If there is no book here, add the book item
function onEnter(user, room) {

    roundNow = UtilGetRoundNumber();

    nextSpawnRound = lastSpawnRound + UtilGetSecondsToRounds(30);
    if ( lastSpawnRound > 0 && roundNow < nextSpawnRound ) {
        return;
    }

    allItems = room.GetItems();

    bookExists = false;
    for ( i=0; i<allItems.length; i++ ) {
        if ( allItems[i].ItemId() == 10 ) {
            bookExists = true;
            return;
        }
    }

    if ( !bookExists ) {
        if ( !user.HasQuest("6-return") && !user.HasItemId(10) ) {
            room.SpawnItem(10, false);
            lastSpawnRound = roundNow;
        }
    }
}
