

ropeAvailableRound = 0;
downRoom = 873;

const chasm = ["chasm", "cliff", "edge", "down", "gorge"];
const bridge = ["bridge"];
const verbs = ["climb", "down", "descend"];

function onCommand_look(rest, user, room) {

    roundNow = UtilGetRoundNumber();

    matches = UtilFindMatchIn(rest, bridge);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "The bridge is uncrossable now. Before you is a large chasm that disappears into darkness.");
        if (roundNow >= ropeAvailableRound) {
            SendUserMessage(user.UserId(), "Someone has tied a rope to a tree here. They must have climbed down.");
        }

        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the broken bridge.");   
        return true;
    }

    matches = UtilFindMatchIn(rest, chasm);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "The cliff is too steep to climb down. You'd need a rope or something to descend.");
        if (roundNow >= ropeAvailableRound) {
            SendUserMessage(user.UserId(), "Someone has tied a rope to a tree here. They must have climbed down.");
        }
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" looks down into the chasm.");   
        return true;
    }

    return false;
}

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();

    climbDown = false;

    if (roundNow < ropeAvailableRound) {

        SendUserMessage(user.UserId(), "You climb down the rope into the chasm.");
        
        user.MoveRoom(downRoom);
        user.Command("look");

        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" pulls out a rope, ties one end to a tree and descends into the chasm.");
        climbDown = true;

    } else {
        
        if ( !user.HasItemId(23) ) {
            SendUserMessage(user.UserId(), "There's really no way down into the chasm without assistance or the right tool.");
            SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" tempts fate by getting a little too close to the edge.");   
            return true;
        }

        SendUserMessage(user.UserId(), "You pull out your rope, tie one end to a tree and descend into the chasm.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" pulls out a rope, ties one end to a tree and descends into the chasm.");

        user.MoveRoom(downRoom);
        user.Command("look");
        
        ropeAvailableRound = roundNow + UtilGetMinutesToRounds(5);

        climbDown = true;
    }


    if ( climbDown ) {

        partyMembers = user.GetPartyMembers();

        for( i = 0; i < partyMembers.length; i++ ) {
            
            a = partyMembers[i];

            if (  a.UserId() == user.UserId() ) {
                continue;
            }

            if ( a.GetRoomId() == room.RoomId() ) {

                if ( a.UserId() > 0 ) {
                    
                    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" pulls out a rope, ties one end to a tree and descends into the chasm.");

                    SendUserMessage(a.UserId(), "You follow "+user.GetCharacterName(true)+" down the rope.");
                    SendRoomMessage(room.RoomId(), a.GetCharacterName(true)+" climbs down the rope.", a.UserId());
                } else {
                    SendRoomMessage(room.RoomId(), a.GetCharacterName(true)+" climbs down the rope.");
                }

                a.MoveRoom(downRoom);
                a.Command("look");
            }
        }

    }

    return true;
}