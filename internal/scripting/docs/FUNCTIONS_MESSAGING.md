# Messaging Specific Functions

Globally available messaging functions.

- [Messaging Specific Functions](#messaging-specific-functions)
  - [SendBroadcast(message string)](#sendbroadcastmessage-string)
  - [SendUserMessage(userId int, message string)](#sendusermessageuserid-int-message-string)
  - [SendRoomMessage(roomId int, message string \[, excludeUserIds int\])](#sendroommessageroomid-int-message-string--excludeuserids-int)
  - [SendRoomExitsMessage(roomId int, message string, isQuiet bool  \[, excludeUserIds int\])](#sendroomexitsmessageroomid-int-message-string-isquiet-bool---excludeuserids-int)

## [SendBroadcast(message string)](/internal/scripting/messaging_func.go)
Sends a message to everyone on the server

|  Argument | Explanation |
| --- | --- |
| message | The message to send. |

## [SendUserMessage(userId int, message string)](/internal/scripting/messaging_func.go)
Sends a message to the userId specified

|  Argument | Explanation |
| --- | --- |
| userId | The userId who should receive the message. |
| message | The message to send. |

## [SendRoomMessage(roomId int, message string [, excludeUserIds int])](/internal/scripting/messaging_func.go)
Sends a message to all users in the roomId specified

|  Argument | Explanation |
| --- | --- |
| roomId | The roomId to transmit the message to. |
| message | The message to send. |
| excludeUserIds | One or more comma separated userIds to exclude from receiving the message. |

## [SendRoomExitsMessage(roomId int, message string, isQuiet bool  [, excludeUserIds int])](/internal/scripting/messaging_func.go)
Sends a message to all rooms with an exit leading to this room

|  Argument | Explanation |
| --- | --- |
| roomId | The roomId to transmit the message to. |
| message | The message to send. |
| isQuiet | If true, only those with superior "hearing" will see it. |
| excludeUserIds | One or more comma separated userIds to exclude from receiving the message. |


