

const allowed_commands = ["help", "broadcast"];
const teach_commands = ["look", "look orb", "look", "look east", "east"];
const teacherMobId = 57;
const teacherName = "Orb of Vision";

var commandNow = 0; // Which command they are on
var canGoEast = false;




// Generic Command Handler
function onCommand(cmd, rest, user, room) {


    ignoreCommand = false;

    teacherMob = getTeacher(room);

    // Make sure they are only doing stuff that's allowed.

    if ( cmd == "east" && !canGoEast ) {
        teacherMob.Command("say Not so hasty! Lets finish the basics before you leave this area.");
        ignoreCommand = true;
    }

    if ( rest.length > 0 ) {
        cmd += ' ' + rest;
    }

    if ( teach_commands[commandNow] == cmd ) {
        
        teacherMob.Command("say Good job!", 4);

        if ( cmd == "look orb" ) {
            teacherMob.Command('say As you can see, looking at me shows you a description and some information about what I\'m carrying.', 4);
        }

        if ( cmd == "look east" ) {
            teacherMob.Command('say Looking into exits like that shows you what (or who) is in a room before you visit it.', 4);
            teacherMob.Command('say Later when you find objects, you can look at them in the same manner.', 4);
            teacherMob.Command('say It\'s always worth trying to look at something you\'re curious about, just in case.', 4);
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
            teacherMob.Command('say The first thing you need to learn is how to inspect your surroundings', 4);
            teacherMob.Command('say type <ansi fg="command">look</ansi> and hit enter to see a description of the area you are in.', 4);
            break;
        case 1:
            teacherMob.Command('say You can also look at creatures or people in the room.', 4);
            teacherMob.Command('say type <ansi fg="command">look orb</ansi> to look at me, ' + teacherMob.GetCharacterName(true) + '.', 4);
            break;
        case 2:
            teacherMob.Command('say Try the <ansi fg="command">look</ansi> command again, but this time, pay attention to any Exits.', 4);
            break;
        case 3:
            teacherMob.Command('say Did you notice there is an exit to the <ansi fg="exit">east</ansi>?', 4);
            teacherMob.Command('say type <ansi fg="command">look east</ansi> to look into the <ansi fg="exit">east</ansi> room.', 4);
            break;
        case 4:
            canGoEast = true;
        default:
            teacherMob.Command('say It\'s time to move on to the next thing you\'ll learn about.', 4);
            teacherMob.Command('say type <ansi fg="command">east</ansi> to travel through the <ansi fg="command">east</ansi> exit.', 4);
            break;
    }
    
    return ignoreCommand;
}




// If there is no book here, add the book item
function onEnter(user, room) {
    teacherMob = getTeacher(room);
    canGoEast = false;
    commandNow = 0;

    teacherMob.Command('say Welcome to the Newbie School!', 4);
    teacherMob.Command('say I\'ll give you some tips to help you get started.', 4);
    teacherMob.Command('say In this area you\'ll learn the basics of inspecting your environment with the <ansi fg="command">look</ansi> command.', 4);
    teacherMob.Command('say type <ansi fg="command">look</ansi> and hit enter to see a description of the area you are in.', 8);
}



function onExit(user , room) {
    // Destroy the guide (cleanup)
    destroyTeacher(room);
}



function onLoad(room) {
    getTeacher(room);
    canGoEast = false;
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