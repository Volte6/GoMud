
function onLost(user, item, room) {
    SendUserMessage(user.UserId(), "You feel disappointment at the loss.");
}

function onFound(user, item, room) {
    SendUserMessage(user.UserId(), "This feels... important.");
}
