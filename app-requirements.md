Server listens to user signups/ logins, chatroom messages/creation, only the admin can add users,
users can req to join a room, chatrooms should terminate when all members but one leave.
if an admin is the only one left in a room, it doesn't terminate, but the room will terminate if they leave to join another room.
if last member is not admin, they will be kicked out of the room and given a noti that room was terminated.
if admin leaves, admin role transfers to next member by join order.


(https://github.com/coder/websocket)
cli tool lib (https://cli.urfave.org/)

## Ideas: 
text colors, ascii sticker like messages, file uploads

## Monitoring & Logging:
user is given a log of all prev active chatrooms should they want to join back.

Implement logging to monitor who joins/leaves chatrooms, when chatrooms are created/terminated, and other critical actions. This will help debug issues and detect any suspicious behavior.

## Rate Limiting for message sending:
To prevent spamming or flooding in chatrooms, consider adding rate limits on the number of messages a user can send in a given period.


## Kick functionality
Add "Kick" functionality for Admins: Allow admins to kick users from chatrooms instead of just adding/inviting users. This will give admins more control over chatroom management.
kicked users can req to join back or be added by admin, banned users cannot, and chatroom will not be displayed to them on search until they are unbanned.

## Custom Error Handling
Provide informative feedback to users when they perform invalid actions (e.g., trying to join a non-existent room or sending messages without being in a room). TODOS: - IMPORTANT: remove user-chatroom associations from db, should be done when user is not banned, but hasn't joined in 30 days

## TODOS:
- remove old notifications
- for the future, look into errors raised from the db and raise the proper messages
- give the ability for owner to make admins
- handle owner / last admin leaving chatroom (no admins left in a chatroom) 
- (Chore) make the json tags in the backend models snake case, make sure they're the same as the ones in the client
- Right now the deployment flow is passing the SERVER_URL into the dockerfile, check how that works for other methods

- Hide Chatrooms that user is banned from
- Join chatroom view, for private chatrooms, the id should be provided in the invite noti, user enters id and joins only if they are invited and are not banned
- handle client errors more gracefully

- After deploying, add Github actions for compiling binaries and displaying them in releases
- Solve the issue where the token isn't read on a different device, it is still saved however,
- Fix issue where user who just registered can't perform actions like creating chatrooms as server cannot detect new user for some reason, test for other actions
- Implement unique field constraints for models such as chatroom titles, user names
- When starting the app logged in, the main chat model displays all chatrooms without pagination, not the case when starting not logged in

