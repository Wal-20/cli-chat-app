cli tool lib (https://cli.urfave.org/)

## Ideas: 
text colors, ascii sticker like messages, file uploads

## Monitoring & Logging:
user is given a log of all prev active chatrooms should they want to join back.

Implement logging to monitor who joins/leaves chatrooms, when chatrooms are created/terminated, and other critical actions. This will help debug issues and detect any suspicious behavior.

## Rate Limiting for message sending:
To prevent spamming or flooding in chatrooms, consider adding rate limits on the number of messages a user can send in a given period.


## Custom Error Handling
Provide informative feedback to users when they perform invalid actions (e.g., trying to join a non-existent room or sending messages without being in a room). TODOS: - IMPORTANT: remove user-chatroom associations from db, should be done when user is not banned, but hasn't joined in 30 days

## TODOS:
- remove old notifications using cron job on server deployment
- for the future, look into errors raised from the db and raise the proper messages
- (Chore) make the json tags in the backend models snake case, make sure they're the same as the ones in the client
- Right now the deployment flow is passing the SERVER_URL into the dockerfile, check how that works for other methods

- handle client errors more gracefully
- Find a way to make binaries smaller

- After deploying, add Github actions for compiling binaries and displaying them in releases
- Solve the issue where the token isn't read on a different device, it is still saved however,
- Implement unique field constraints for models such as chatroom titles, user names
- When starting the app logged in, the main chat model displays all chatrooms without pagination, not the case when starting not logged in

