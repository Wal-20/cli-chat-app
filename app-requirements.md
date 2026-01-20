cli tool lib (https://cli.urfave.org/)

## Ideas: 
- text colors (among other client specified config), ascii sticker like messages, file uploads
- File Previews (Inline)
For example:
Images → show ASCII preview, store S3 upload link
PDFs → show metadata, store S3 upload link
Text files → show first 10 lines

- Encryption (E2E Mode): You can add optional end-to-end encryption; server only relays ciphertext, keys exchanged via Curve25519, messages encrypted with AES-GCM or ChaCha20

- Admin screen with:
users joined/left
mutes
kicks
bans
role changes

separate api to encrypt and decrypt with several algorithms and complexities, served as a string or a file, the actual function that will encrypt the data will run as a command line utility, it's just the output that is served with the api

## Monitoring & Logging:
user is given a log of all prev active chatrooms should they want to join back.

Implement logging to monitor who joins/leaves chatrooms, when chatrooms are created/terminated, and other critical actions. This will help debug issues and detect any suspicious behavior.

## Rate Limiting for message sending:
To prevent spamming or flooding in chatrooms, consider adding rate limits on the number of messages a user can send in a given period.

## TODOS:
- remove old notifications cron job (needs testing)
- for the future, look into errors raised from the db and raise the proper messages
- (Chore) make the json tags in the backend models snake case, make sure they're the same as the ones in the client
- Right now the deployment flow is passing the SERVER_URL into the dockerfile, check how that works for other methods

- After deploying, add Github actions for pulling newest code to the vps and compiling binaries and displaying them in releases
- Implement unique field constraints for models such as chatroom titles, user names
- When starting the app logged in, the main chat model displays all chatrooms without pagination, not the case when starting not logged in

- implement a stack for wsEvents and sort by priority, typing first, then user left/joined, display events as two rows, first the typing events, then the left/joined events, the stack should be on the wsserver, client only sends wsEvent to update the stack?
