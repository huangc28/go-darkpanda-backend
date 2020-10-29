## TODOs

- [x] Get inquiries. Query condition should exclude expired records. We should add `expired_at` column in `lobby_users` table and remove `expired_at` in `lobby_users` table.
- [x] Cancel inquiry. 
- [x] User revert inquiry from chatting. Soft both delete user's presence in that chatroom and the chatroom itself. If is a male user, rejoining lobby if inquiry isn't expired.
- [] Reverting inquiry should emit `quit` message to members in the channel.
- [] When service provider picks up an inquiry, backend should create a private chat document in firestore. Client will then subscribe to this document and the subcollections beneath this document. 

## Bugs

- [] If a service provider is already chatting with the customer, she should not be able to pickup the same inquiry again.
- [x] If an inquiry is at `chatting` status, it should be removed to lobby. Other service providers should not be able to find the inquiry.
- [x] Error #01: sql: transaction has already been committed or rolled back.

Committed / Roll backed transaction should be replaced with normal SQL client. A better solution would be to resolve a whole new DAO Service object on each request to avoid reusing Roll backed / Committed transactions.