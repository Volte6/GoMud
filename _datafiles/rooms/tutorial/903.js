
const allowed_commands = ["help", "broadcast", "look", "status", "inventory", "experience", "conditions", "equip"];
const teach_commands = ["get cap", "equip cap", "portal"];
const teacherMobId = 57;
const teacherName = "Orb of Graduation";
const capItemId = 20043;

var commandNow = 0; // Which command they are on



// Generic Command Handler
function onCommand(cmd, rest, user, room) {
    
    ignoreCommand = false;

    teacherMob = getTeacher(room);

    fullCommand = cmd;
    if ( rest.length > 0 ) {
        fullCommand = cmd + ' ' + rest;
    }

    if ( commandNow >= 2 ) {
        return false;
    }
    
    if ( teach_commands[commandNow] == fullCommand ) {
        
        if ( fullCommand == "equip cap" ) {
            teacherMob.Command("say Good job!");
        } else {
            teacherMob.Command("say Good job! You earned it!");
        }

        commandNow++;

    } else {

        if ( allowed_commands.includes(cmd) || teach_commands.slice(0, commandNow).includes(cmd) ) {
            return false;
        }
        
        ignoreCommand = true;
    }

    switch (commandNow) {
        case 0:

            teacherMob.Command('emote gestures to the <ansi fg="item">graduation cap</ansi> on the ground.')
            teacherMob.Command('say type <ansi fg="command">get cap</ansi> to pick up the <ansi fg="item">graduation cap</ansi>.');
            break;
        case 1:

            teacherMob.Command('say Go ahead and wear the <ansi fg="item">graduation cap</ansi> by typing <ansi fg="command">equip cap</ansi>.');
            break;
        case 2:

            teacherMob.Command('say It\'s time to say goodbye');
            teacherMob.Command('say I\ll summon a portal to send you to the heart of Frostfang city, where your adventure begins.');

            exits = room.GetExits();
            if ( !exits['portal'] ) {
                teacherMob.Command('emote glows intensely, and a ' + UtilApplyColorPattern('swirling portal', 'pink') + ' appears!');
                room.AddTemporaryExit('swirling portal', ':pink', 1, 9000);
            }

            teacherMob.Command('say Enter the portal when you are ready.');
            
            break;
        default:
            break;
    }
    
    return ignoreCommand;
}




// If there is no book here, add the book item
function onEnter(user, room) {
    
    teacherMob = getTeacher(room);
    clearGroundItems(room);
    
    sendWorkingCommands(user);

    itm = CreateItem(capItemId);
    teacherMob.GiveItem(itm);

    teacherMob.Command('emote appears in a ' + UtilApplyColorPattern("flash of light!", "glowing"));

    teacherMob.Command('say Congratulation on getting to the end of the training course!');
    teacherMob.Command('drop cap');
    teacherMob.Command('emote gestures to the <ansi fg="item">graduation cap</ansi> on the ground.', 15)
    teacherMob.Command('say type <ansi fg="command">get cap</ansi> to pick up the <ansi fg="item">graduation cap</ansi>.', 15);

}



function onExit(user , room) {
    // Destroy the guide (cleanup)
    destroyTeacher(room);
    
    canGoSouth = false;
    commandNow = 0;
}



function onLoad(room) {
    canGoSouth = false;
    commandNow = 0;
}


function getTeacher(room) {

    var mobActor = null;

    mobIds = room.GetMobs();
    
    for ( i in mobIds ) {
        mobActor = GetMob(mobIds[i]);
        if ( mobActor.MobTypeId() == teacherMobId ) {
            return mobActor;
        }
    }

    mobActor = room.SpawnMob(teacherMobId);
    mobActor.SetCharacterName(teacherName);

    return mobActor;
}

function destroyTeacher(room) {

    var mobActor = null;

    mobIds = room.GetMobs();
    
    for ( i in mobIds ) {
        mobActor = GetMob(mobIds[i]);
        if ( mobActor.MobTypeId() == teacherMobId ) {
            mobActor.Command(`suicide vanish`);
        }
    }
}


function sendWorkingCommands(user) {

    ac = [];
    unlockedCommands = teach_commands.slice(0, commandNow);

    for (var i in allowed_commands ) {
        ac.push(allowed_commands[i]);
    }
    
    for (var i in unlockedCommands ) {
        ac.push(unlockedCommands[i]);
    }
    
    user.SendText("");
    user.SendText("");
    user.SendText('    <ansi fg="red">NOTE:</ansi> Most commands have been <ansi fg="203">DISABLED</ansi> and <ansi fg="203">WILL NOT WORK</ansi> until you <ansi fg="51">COMPLETE THIS TUTORIAL</ansi>!');
    //user.SendText('          The commands currently available are: <ansi fg="command">'+ac.join('</ansi>, <ansi fg="command">')+'</ansi>');
    user.SendText("");
    user.SendText("");

}

function clearGroundItems(room) {

    allGroundItems = room.GetItems();
    for ( i in allGroundItems ) {
        room.DestroyItem(allGroundItems[i]);
    }

}