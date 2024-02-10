# Messaging Specific Functions

---

[SendBroadcast(message string)](messaging_func.go) - _Sends a message to everyone on the server_

|  Argument | Explanation |
| --- | --- |
| message | The message to send. |

---

[SendUserMessage(userId int, message string)](messaging_func.go) - _Sends a message to the userId specified_

|  Argument | Explanation |
| --- | --- |
| userId | The userId who should receive the message. |
| message | The message to send. |

---

[SendRoomMessage(roomId int, message string, [, excludeUserIds int])](messaging_func.go) - _Sends a message to all users in the roomId specified_

Note: If this is in a function for an event a user triggered, they will automatically be excluded.

|  Argument | Explanation |
| --- | --- |
| roomId | The roomId to transmit the message to. |
| message | The message to send. |
| excludeUserIds | One or more comma separated userIds to exclude from receiving the message. |

---
