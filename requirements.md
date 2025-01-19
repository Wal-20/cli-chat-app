

CLI CHAT THINGY

server listens to user signups/ logins, chatroom messages/creation, only the admin can add users,
users can req to join a room, chatrooms should terminate when all members but one leave.
if an admin is the only one left in a room, it doesn't terminate, but the room will terminate if they leave to join another room.
if last member is not admin, they will be kicked out of the room and given a noti that room was terminated.
if admin leaves, admin role transfers to next member by join order.

admin invites members through id or username

user is given a log of all prev active chatrooms should they want to join back.

TUI (terminal user interface) lib (https://earthly.dev/blog/tui-app-with-go/)
cli tool lib (https://cli.urfave.org/)

look at https://github.com/charmbracelet/bubbletea


SCHEMAS:

user: id, username, password, chatrooms 
chatrooms: id, users, messages, adminId, maxUserCount
message: id, senderId, timestamp, chatroomId
logEntry: id, 

(e.g., index chatroom IDs for fast lookup).


user <--MANY_MANY--> chatroom
chatroom <--ONE_MANY--> message
user <--ONE_MANY--> message


HOW WILL A USER JOIN A ROOM?


ideas: 
text colors, ascii sticker like messages, file uploads


Monitoring & Logging:

Implement logging to monitor who joins/leaves chatrooms, when chatrooms are created/terminated, and other critical actions. This will help debug issues and detect any suspicious behavior.
Rate Limiting for Message Sending:

Rate Limiting for message sending:

To prevent spamming or flooding in chatrooms, consider adding rate limits on the number of messages a user can send in a given period.


Add "Kick" functionality for Admins: Allow admins to kick users from chatrooms instead of just adding/inviting users. This will give admins more control over chatroom management.
kicked users can req to join back or be added by admin, banned users cannot, and chatroom will not be displayed to them on search until they are unbanned.

Optional Public Rooms: In the future, you may want to allow for public chatrooms that anyone can join without needing an invite.

Custom Error Messages: Provide informative feedback to users when they perform invalid actions (e.g., trying to join a non-existent room or sending messages without being in a room).



TODOS:
- secret key user has to enter to join a room, we'll cross this bridge when we get there.
- logging for chatroom members, join time, kicked / banned status
