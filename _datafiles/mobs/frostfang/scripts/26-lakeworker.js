
SAID_STUFF = false;
WALK_DIRECTION = 1;
WALK_POSITION = 0;
WALK_PATH = [362, // old dock
            333,
            332,
            331,
            330,
            329,
            328,
            327,
            326,
            325,
            324,
            323,
            322,
            321,
            320,
            319] // crashing waves leading to the rocky island


const boatNouns = ["boats", "oars", "ships", "paddles"];
const crashNouns = ["rocks", "crash", "choppy", "water", "waves"];

function onAsk(mob, room, eventDetails) {

    roomId = room.RoomId();
    
    match = UtilFindMatchIn(eventDetails.askText, boatNouns);
    if ( match.found ) {

        if ( roomId == 319 ) {
            
            mob.Command("say I hit those rocks just over there and lost all of our oars.");
            mob.Command("emote points to the northwest.");

        } else {
            if ( WALK_POSITION >= 5 ) {
                mob.Command("say We lost the oars to the boats when the choppy water caused me to crash against some rocks to the west of here.");
            } else {
                mob.Command("say We lost the oars to the boats when the choppy water caused me to crash against some rocks in the southwest part of the lake.");
            }
        }
        return;
    }

    match = UtilFindMatchIn(eventDetails.askText, crashNouns);
    if ( match.found ) {


        if ( roomId == 319 ) {
            
            mob.Command("say I hit those rocks just over there and lost all of our oars.");
            mob.Command("emote points to the northwest.");

        } else {
            if ( WALK_POSITION >= 5 ) {
                mob.Command("say Just a little west of here are some rocky islands.");
                if ( WALK_DIRECTION > 0 ) {
                    mob.Command("say I'm headed there now to see if there's any trash washed ashore.");
                }else {
                    mob.Command("say I'm just coming back from cleaning up trash from there.");
                }
            } else {
                mob.Command("say In the southwest part of the lake are some rocky islands.");
                mob.Command("say I visit there every so often to clean up trash that washes ashore.");

                if ( WALK_DIRECTION > 0 ) {
                    mob.Command("say I'm heading there soon to see if there's any trash washed ashore.");
                }else {
                    mob.Command("say I'm just coming back from cleaning up trash from there.");
                }
            }
        }
        return;

    }

}

function onGive(mob, room, eventDetails) {

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 10016) {
            mob.Command("look !"+String(eventDetails.item.ItemId))
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
            return true;
        }

        mob.Command("say Thanks, but my days of rowing are over. I'm just a lowly lake worker now.");
        mob.Command("drop !"+String(eventDetails.item.ItemId))
        return true;
    }

}


// Invoked once every round if mob is idle
function onIdle(mob, room) {


    if ( mob.GetRoomId() == 319 ) {

        if ( !SAID_STUFF ) {
            mob.Command("emote squints and peers towards a rocky island in the lake to the northwest.");
            mob.Command("emote mutters to himself.");
            if ( UtilDiceRoll(1, 2) == 1 ) {
                mob.Command("say Ever since I crashed my boat on those rocks, I've been demoted to cleaning up the lakeshore.");
            }

            SAID_STUFF = true;
            return true;
        }

        SAID_STUFF = false; // reset

    } else if ( UtilDiceRoll(1, 2) > 1 ) {
        return true;
    }



    if ( WALK_POSITION < 0 ) {
        WALK_POSITION = 0;
    } else if ( WALK_POSITION > WALK_PATH.length - 1) {
        WALK_POSITION = WALK_PATH.length - 1;
    }

    roomNow = WALK_PATH[WALK_POSITION];

    if ( roomNow != mob.GetRoomId() ) {
        
        WALK_POSITION = 0;
        WALK_DIRECTION = 1;
        mob.MoveRoom(WALK_PATH[WALK_POSITION]);

    } else {

        if ( WALK_POSITION >= WALK_PATH.length -1 ) {
            WALK_DIRECTION = -1;
        }
        if ( WALK_POSITION < 0 ) {
            WALK_DIRECTION = 1;
        }

        WALK_POSITION += WALK_DIRECTION;

        exitList = room.GetExits();
        for (var key in exitList) {
            if ( exitList[key].RoomId == WALK_PATH[WALK_POSITION] ) {
                mob.Command( exitList[key].Name );
                
            }
        }
        
    }

    return true;
}

